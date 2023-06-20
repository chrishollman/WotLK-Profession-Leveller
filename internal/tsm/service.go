package tsm

import (
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	rhttp "github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"

	"github.com/chrishollman/WotLK-Profession-Leveller/internal/cache"
)

const (
	// URLs
	tsmAuthURL string = "https://auth.tradeskillmaster.com/oauth2/token"
	tsmBaseURL string = "https://pricing-api.tradeskillmaster.com/ah/"

	// Auth request
	authClientID  string = "c260f00d-1071-409a-992f-dda2e5498536"
	authGrantType string = "api_token"
	authScope     string = "app:realm-api app:pricing-api"
)

type TSMItemRes struct {
	ItemID      int `json:"itemId"`
	MinBuyout   int `json:"minBuyout"`   // Current buyout value
	MarketValue int `json:"marketValue"` // Average value over last 2 weeks
	Historical  int `json:"historical"`  // Average value over last 2 months
	NumAuctions int `json:"numAuctions"` // Number of auctions live
}

// Houses the HTTP client used to receive data from the TSM api, and a local caching service to cache those calls.
type TSMService struct {
	blacklist []int
	cfg       *tsmcfg
	cache     *bigcache.BigCache
	client    *rhttp.Client
	logger    *zap.SugaredLogger
}

type tsmcfg struct {
	apiKey    string
	authToken string
	expiry    int
}

type tsmAuthReq struct {
	ClientID  string `json:"client_id"`
	GrantType string `json:"grant_type"`
	Scope     string `json:"scope"`
	Token     string `json:"token"`
}

type tsmAuthRes struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expiry       int    `json:"expires_in"`
}

func NewTSMService(apiKey string, logger *zap.SugaredLogger) *TSMService {
	return &TSMService{
		blacklist: []int{
			12662, // demonic rune
			18240, // ogre tannin
		},
		cfg:    &tsmcfg{apiKey: apiKey},
		cache:  cache.NewBigCache(),
		client: rhttp.NewClient(),
		logger: logger,
	}
}

func (ts *TSMService) AuthTicker() {
	ts.auth()

	expiryTicker := time.NewTicker(time.Duration(int(ts.cfg.expiry)) * time.Second)

	go func() {
		for {
			select {
			case <-expiryTicker.C:
				ts.auth()
			}
		}
	}()
}

// Returns the current minBuyout for a provided item
func (i *TSMItemRes) Price(pricingType string) int {
	switch {
	case strings.EqualFold(pricingType, "marketvalue"):
		return i.MarketValue
	case strings.EqualFold(pricingType, "historical"):
		return i.Historical
	case strings.EqualFold(pricingType, "minbuyout"):
		return i.MinBuyout
	default:
		return i.MarketValue
	}
}
