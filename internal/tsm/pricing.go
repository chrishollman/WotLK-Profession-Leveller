package tsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/allegro/bigcache/v3"
	rhttp "github.com/hashicorp/go-retryablehttp"
)

var (
	ErrIsBlacklisted error = errors.New("item is blacklisted")
)

func (ts *TSMService) Preheat(ahID int) error {
	// Check cache to see if entry already exists
	_, err := ts.cache.Get(fmt.Sprintf("%v", ahID))

	switch {
	case err == nil:
		// continue
	case errors.Is(err, bigcache.ErrEntryNotFound):
		err = ts.getFull(ahID)
		if err != nil {
			return err
		}
	}

	return err
}

func (ts *TSMService) GetPrice(ahID int, itemID int) (*TSMItemRes, error) {
	var key = fmt.Sprintf("%v-%v", ahID, itemID)

	// Check blacklist
	if ts.isBlacklisted(itemID) {
		return nil, ErrIsBlacklisted
	}

	// Get from cache
	item, err := ts.getFromCache(key)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (ts *TSMService) getFromCache(key string) (*TSMItemRes, error) {
	res := &TSMItemRes{}

	// Check to see if the data exists in the cache
	bytes, err := ts.cache.Get(key)
	switch {
	case err == nil:
		// continue
	case errors.Is(err, bigcache.ErrEntryNotFound):
		return nil, fmt.Errorf("cache miss for %v", key)
	default:
		return nil, fmt.Errorf("unexpected error in cache check: %v", err)
	}

	err = json.Unmarshal(bytes, res)
	if err != nil {
		return nil, fmt.Errorf("couldn't read item from cache: %v", err)
	}

	return res, nil
}

func (ts *TSMService) getFull(ahID int) error {
	reqURL := fmt.Sprintf("%v%v", tsmBaseURL, ahID)

	req, err := rhttp.NewRequest("GET", reqURL, nil)
	if err != nil {
		ts.logger.Fatalf("couldn't create request for AH data: %v", err)
	}
	req.Header.Set("Authorization", ts.getBearer())

	// Send
	res, err := ts.client.Do(req)
	if err != nil {
		ts.logger.Fatalf("couldn't send request for AH data: %v", err)
	}

	// Parse
	var resParsed []TSMItemRes
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &resParsed)
	if err != nil {
		return err
	}

	// Save lookup key
	ts.cache.Set(fmt.Sprintf("%v", ahID), []byte(`1`))

	// Save data
	for _, v := range resParsed {
		key := fmt.Sprintf("%v-%v", ahID, v.ItemID)
		tmp, _ := json.Marshal(v)
		err = ts.cache.Set(key, tmp)
		if err != nil {
			ts.logger.Error("couldn't set item with key %v", key)
		}
	}

	return err
}

// Unused - Fetch a single item from the AH. Due to API limitations (500 single requests vs 100 server requests) it
// seems better to simply fetch the entire AH instead of a singular item.
func (ts *TSMService) getItem(ahID int, itemID int) error {
	reqURL := fmt.Sprintf("%v%v/item/%v", tsmBaseURL, ahID, itemID)

	req, err := rhttp.NewRequest("GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("couldn't create request for AH item data: %w", err)
	}
	req.Header.Set("Authorization", ts.getBearer())

	// Send
	res, err := ts.client.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't send request for AH item data: %w", err)
	}

	// Parse
	var resParsed *TSMItemRes
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("couldn't parse request for AH item data: %w", err)
	}

	err = json.Unmarshal(body, resParsed)
	if err != nil {
		return fmt.Errorf("couldn't unmarshal response into item struct: %w", err)
	}

	// Save
	tmp, _ := json.Marshal(resParsed)
	err = ts.cache.Set(fmt.Sprintf("%v-%v", ahID, resParsed.ItemID), tmp)
	if err != nil {
		return fmt.Errorf("couldn't save item to cache: %w", err)
	}

	return nil
}

func (ts *TSMService) isBlacklisted(itemID int) bool {
	for _, v := range ts.blacklist {
		if v == itemID {
			return true
		}
	}
	return false
}
