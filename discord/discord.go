package discord

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kehiy/RoboPac/client"
	"github.com/kehiy/RoboPac/config"
	"github.com/kehiy/RoboPac/wallet"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pactus-project/pactus/crypto"
	"github.com/pactus-project/pactus/util"
	pactus "github.com/pactus-project/pactus/www/grpc/gen/go"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Bot struct {
	discordSession *discordgo.Session
	faucetWallet   *wallet.Wallet
	cfg            *config.Config
	store          *SafeStore

	cm *client.Mgr
}

// guildID: "795592769300987944"

func Start(cfg *config.Config, w *wallet.Wallet, ss *SafeStore) (*Bot, error) {
	cm := client.NewClientMgr()

	for _, s := range cfg.Servers {
		c, err := client.NewClient(s)
		if err != nil {
			log.Printf("unable to create client at: %s. err: %s", s, err)
		} else {
			log.Printf("adding client at: %s", s)
			cm.AddClient(s, c)
		}
	}
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Printf("error creating Discord session: %v", err)
		return nil, err
	}
	bot := &Bot{cfg: cfg, discordSession: dg, faucetWallet: w, store: ss, cm: cm}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(bot.messageHandler)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Printf("error opening connection: %v", err)
		return nil, err
	}

	return bot, nil
}

func (b *Bot) Stop() error {
	return b.discordSession.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func (b *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	p := message.NewPrinter(language.English)
	// log.Printf("received message: %v\n", m.Content)

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.ToLower(m.Content) == "help" {
		// msg := "You can request the faucet by sending your wallet address, e.g tpc1pxl333elgnrdtk0kjpjdvky44yu62x0cwupnpjl"
		// s.ChannelMessageSendReply(m.ChannelID, msg)
		help(s, m)
		return
	}

	if strings.ToLower(m.Content) == "network" {
		msg := b.networkInfo()
		_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		return
	}

	if strings.ToLower(m.Content) == "address" {
		msg := fmt.Sprintf("Faucet address is: %v", b.cfg.FaucetAddress)
		_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		return
	}

	// If the message is "balance" reply with "available faucet balance"
	if strings.ToLower(m.Content) == "balance" {
		balance := b.faucetWallet.GetBalance()
		v, d := b.store.GetDistribution()
		msg := p.Sprintf("Available faucet balance is %.4f PACs\n", balance.Available)
		msg += p.Sprintf("A total of %.4f PACs has been distributed to %d validators\n", d, v)
		_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		return
	}

	if strings.Contains(strings.ToLower(m.Content), "peer-info") {
		trimmedPrefix := strings.TrimPrefix(strings.ToLower(m.Content), "peer-info")
		trimmedAddress := strings.Trim(trimmedPrefix, " ")

		peerInfo, err := b.GetPeerInfo(trimmedAddress)
		if err != nil {
			msg := p.Sprintf("An error occurred %v\n", err)
			_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
			return
		}

		peerID, err := peer.IDFromBytes(peerInfo.PeerId)
		if err != nil {
			msg := p.Sprintf("An error occurred %v\n", err)
			_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
			return
		}

		msg := p.Sprintf("Peer info ,\n")
		msg += p.Sprintf("Peer ID = %v\n", peerID)
		msg += p.Sprintf("IP address = %v\n", peerInfo.Address)
		msg += p.Sprintf("Agent =  %v\n", peerInfo.Agent)
		msg += p.Sprintf("Moniker = %v\n", peerInfo.Moniker)
		_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		return
	}

	if strings.Contains(strings.ToLower(m.Content), "faucet") {
		trimmedPrefix := strings.TrimPrefix(strings.ToLower(m.Content), "faucet")
		// faucet message must contain address/pubkey
		trimmedAddress := strings.Trim(trimmedPrefix, " ")
		peerID, pubKey, isValid, msg := b.validateInfo(trimmedAddress, m.Author.ID)

		msg = fmt.Sprintf("%v\ndiscord: %v\naddress: %v",
			msg, m.Author.Username, trimmedAddress)

		if !isValid {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
			return
		}

		if pubKey != "" {
			// check available balance
			balance := b.faucetWallet.GetBalance()
			if balance.Available < b.cfg.FaucetAmount {
				_, _ = s.ChannelMessageSendReply(m.ChannelID, "Insufficient faucet balance. Try again later.", m.Reference())
				return
			}

			// send faucet
			txHash := b.faucetWallet.BondTransaction(pubKey, trimmedAddress, b.cfg.FaucetAmount)
			if txHash != "" {
				err := b.store.SetData(peerID, trimmedAddress, m.Author.Username, m.Author.ID, b.cfg.FaucetAmount)
				if err != nil {
					log.Printf("error saving faucet information: %v\n", err)
				}
				msg := p.Sprintf("%v  %.4f test PACs is staked to %v successfully!",
					m.Author.Username, b.cfg.FaucetAmount, trimmedAddress)
				_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
			}
		}
	}

	if strings.Contains(strings.ToLower(m.Content), "tx-data") {
		trimmedPrefix := strings.TrimPrefix(strings.ToLower(m.Content), "peer-info")
		trimmedTXHash := strings.Trim(trimmedPrefix, " ")

		data, err := b.cm.GetRandomClient().TransactionData(trimmedTXHash)
		if err != nil {
			msg := p.Sprintf("An error occurred %v\n", err)
			_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
			return
		}

		msg := p.Sprintf("your transaction data:\ndata:%v\nversion:%v\nlockTime:%v\nvalue:%v\nmemo:%v\npubkey:%v\n",
			string(data.Data), data.Version, data.LockTime, data.Value, data.Memo, data.PublicKey)
		_, _ = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		return
	}
}

// help sends a message detailing how to use the bot discord-client side
// nolint.
func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Title: "Pactus Universal Robot",
		URL:   "https://pactus.org",
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://pactus.org",
			IconURL: s.State.User.AvatarURL(""),
			Name:    s.State.User.Username,
		},
		Description: "RoboPac is a robot that provides support and information about the Pactus Blockchain.\n" +
			"To see the faucet account balance, simply type: `balance`\n" +
			"To see the faucet address, simply type: `address`\n" +
			"To get network information, simply type: `network`\n" +
			"To get peer information, simply type: `peer-info [validator address]`\n" +
			"To request faucet for test network: simply post `faucet [validator address]`.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Example of requesting `faucet` ",
				Value: "faucet tpc1pxl333elgnrdtk0kjpjdvky44yu62x0cwupnpjl",
			},
		},
	})
}

func (b *Bot) validateInfo(address, discordID string) (string, string, bool, string) {
	_, err := crypto.AddressFromString(address)
	if err != nil {
		log.Printf("invalid address")
		return "", "", false, "Pactus Universal Robot is unable to handle your request." +
			" If you are requesting testing faucet, supply the valid address."
	}

	// check if the user is existing
	v, exists := b.store.FindDiscordID(discordID)
	if exists {
		return "", "", false, "Sorry. You already received faucet using this address: " + v.ValidatorAddress
	}

	// check if the address exists in the list of validators
	isValidator, err := b.cm.IsValidator(address)
	if err != nil {
		return "", "", false, err.Error()
	}

	if isValidator {
		return "", "", false, "Sorry. Your address is in the list of active validators. You do not need faucet again."
	}

	peerInfo, pub, err := b.cm.GetPeerInfo(address)
	if err != nil {
		return "", "", false, err.Error()
	}
	if pub == nil {
		log.Printf("error getting peer info")
		return "", "", false, "Your node information could not obtained." +
			" Make sure your node is fully synced before requesting the faucet."
	}

	// check if the validator has already been given the faucet
	peerID, err := peer.IDFromBytes(peerInfo.PeerId)
	if err != nil {
		return "", "", false, err.Error()
	}
	if peerID.String() == "" {
		log.Printf("error getting peer id")
		return "", "", false, "Your node information could not obtained." +
			" Make sure your node is fully synced before requesting the faucet."
	}
	v, exists = b.store.GetData(peerID.String())
	if exists || v != nil {
		return "", "", false, "Sorry. You already received faucet using this address: " + v.ValidatorAddress
	}

	// check block height
	// height, err := cl.GetBlockchainHeight()
	// if err != nil {
	// 	log.Printf("error current block height")
	// 	return "", "", false, "The bot cannot establish connection to the blochain network. Try again later."
	// }
	// if (height - peerInfo.Height) > 1080 {
	//	msg := fmt.Sprintf("Your node is not fully synchronised. It is is behind by %v blocks." +
	//		" Make sure that your node is fully synchronised before requesting faucet.", (height - peerInfo.Height))

	// 	log.Printf("peer %s with address %v is not well synced: ", peerInfo.PeerId, address)
	// 	return "", "", false, msg
	// }
	return peerID.String(), pub.String(), true, ""
}

func (b *Bot) networkInfo() string {
	msg := "Pactus is truly decentralised proof of stake blockchain."
	nodes, err := b.cm.GetNetworkInfo()
	if err != nil {
		log.Printf("error establishing connection")
		return msg
	}
	msg += "\nThe following are the currentl statistics:\n"
	msg += fmt.Sprintf("Network started at : %v\n", time.UnixMilli(nodes.StartedAt*1000).Format("02/01/2006, 15:04:05"))
	msg += fmt.Sprintf("Total bytes sent : %v\n", nodes.TotalSentBytes)
	msg += fmt.Sprintf("Total received bytes : %v\n", nodes.TotalReceivedBytes)
	msg += fmt.Sprintf("Number of peer nodes: %v\n", len(nodes.Peers))
	// check block height
	blockchainInfo, err := b.cm.GetBlockchainInfo()
	if err != nil {
		log.Printf("error current block height")
		return msg
	}
	msg += fmt.Sprintf("Block height: %v\n", blockchainInfo.LastBlockHeight)
	msg += fmt.Sprintf("Total power: %.4f PACs\n", util.ChangeToCoin(blockchainInfo.TotalPower))
	msg += fmt.Sprintf("Total committee power: %.4f PACs\n", util.ChangeToCoin(blockchainInfo.CommitteePower))
	msg += fmt.Sprintf("Total validators: %v\n", blockchainInfo.TotalValidators)
	return msg
}

func (b *Bot) GetPeerInfo(address string) (*pactus.PeerInfo, error) {
	_, err := crypto.AddressFromString(address)
	if err != nil {
		log.Printf("invalid address")

		return nil, err
	}

	_, err = b.cm.IsValidator(address)
	if err != nil {
		return nil, err
	}

	peerInfo, _, err := b.cm.GetPeerInfo(address)
	if err != nil {
		return nil, err
	}
	return peerInfo, nil
}
