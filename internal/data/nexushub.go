package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/allegro/bigcache/v3"
	rhttp "github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"

	"github.com/chrishollman/WotLK-Profession-Leveller/internal/cache"
)

const (
	NEXUS_HUB_ITEM_PRICE_BASE_URL string = "https://api.nexushub.co/wow-classic/v1/items/"
)

// Represents the response contract provided by NexusHub
type NexusHubItemRes struct {
	Server        string   `json:"server"`
	ItemID        int      `json:"itemId"`
	Name          string   `json:"name"`
	UniqueName    string   `json:"uniqueName"`
	Icon          string   `json:"icon"`
	Tags          []string `json:"tags"`
	RequiredLevel int      `json:"requiredLevel"`
	ItemLevel     int      `json:"itemLevel"`
	SellPrice     int      `json:"sellPrice"`
	VendorPrice   int      `json:"vendorPrice"`
	ItemLink      string   `json:"itemLink"`
	Tooltip       []struct {
		Label string `json:"label"`
	} `json:"tooltip"`
	Stats struct {
		Current struct {
			MarketValue int `json:"marketValue"`
			MinBuyout   int `json:"minBuyout"`
			Quantity    int `json:"quantity"`
		} `json:"current"`
		Previous struct {
			MarketValue int `json:"marketValue"`
			MinBuyout   int `json:"minBuyout"`
			Quantity    int `json:"quantity"`
		} `json:"previous"`
	} `json:"stats"`
}

// Houses the HTTP client used to receive data from the NexusHub api, and a local caching service to cache those calls.
type NexusHubStore struct {
	cache  *bigcache.BigCache
	client *rhttp.Client
	logger *zap.SugaredLogger
}

func NewNexusHubStore(logger *zap.SugaredLogger) *NexusHubStore {
	return &NexusHubStore{
		cache:  cache.NewBigCache(),
		client: rhttp.NewClient(),
		logger: logger,
	}
}

// Returns the current minBuyout for a provided item
func (n *NexusHubItemRes) Price() int {
	return n.Stats.Current.MinBuyout
}

func (nh *NexusHubStore) getItem(server, faction string, id int) (*NexusHubItemRes, error) {
	data := &NexusHubItemRes{}

	// Check to see if the data exists in the cache
	key := fmt.Sprintf("%v-%v-%v", server, faction, id)
	if bytes, err := nh.cache.Get(key); err == nil {
		jsonErr := json.Unmarshal(bytes, data)
		if jsonErr != nil {
			return nil, err
		}
		return data, nil
	}

	// Not in the cache, query NexusHub
	reqUrl := fmt.Sprintf("%v%v-%v/%v", NEXUS_HUB_ITEM_PRICE_BASE_URL, server, faction, id)
	res, err := nh.client.Get(reqUrl)
	if res.StatusCode != http.StatusOK || err != nil {
		nh.logger.Error(err.Error())
		return nil, err
	}

	// Decode the response from NexusHub
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(data)
	if err != nil {
		nh.logger.Error(fmt.Errorf("cant decode item price: %w", err))
		return nil, err
	}

	// Store the returned data in the cache for future use
	cacheData, _ := json.Marshal(data)
	err = nh.cache.Set(key, cacheData)
	if err != nil {
		nh.logger.Error(fmt.Errorf("cant set item price in the cache: %w", err))
	}

	// Return the result
	return data, nil
}

func (nh *NexusHubStore) GetPrice(server, faction string, id int) (int, error) {
	data, err := nh.getItem(server, faction, id)
	switch {
	case err != nil:
		return 0, err
	case data.Price() == 0:
		return 0, errors.New("pricing unavailable")
	default:
		return data.Price(), nil
	}
}

func (nh *NexusHubStore) GetPriceBatch(server, faction string, id []int) (int, error) {
	var total int
	for _, v := range id {
		tmp, err := nh.getItem(server, faction, v)
		if err != nil {
			return 0, errors.New("unable to get price for batch request")
		}
		total += tmp.Stats.Current.MarketValue
	}
	return total, nil
}
