//usage.go - implements usage admin op API

package radosgw

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

const usageTimeFormat = "2006-01-02 15:04:05"

type UsageCategoryType struct {
	Category      string `json:"category"`
	BytesSent     int64  `json:"bytes_sent"`
	BytesReceived int64  `json:"bytes_received"`
	Ops           int64  `json:"ops"`
	SuccessfulOps int64  `json:"successful_ops"`
}

type UsageEntryType struct {
	User    string `json:"user,omitempty"`
	Buckets []struct {
		Bucket     string              `json:"bucket"`
		Time       string              `json:"time"`
		Epoch      int64               `json:"epoch"`
		Owner      string              `json:"owner"`
		Categories []UsageCategoryType `json:"categories"`
	} `json:"buckets,omitempty"`
}

type UsageSummaryType struct {
	User       string              `json:"user,omitempty"`
	Categories []UsageCategoryType `json:"categories,omitempty"`
	Total      struct {
		BytesSent     int64 `json:"bytes_sent"`
		BytesReceived int64 `json:"bytes_received"`
		Ops           int64 `json:"ops"`
		SuccessfulOps int64 `json:"successful_ops"`
	} `json:"total,omitempty"`
}

type UsageType struct {
	Entries []UsageEntryType   `json:"entries"`
	Summary []UsageSummaryType `json:"summary"`
}

// GetUsage - get the usage info of the radosgw service
//
// PARAMS:
//     - uid: user id string
//     - start: start timestamp
//     - end: end timestamp, not include
//     - showSummary: show summary info or not
//     - showEntries: show entries info or not
// RETURN:
//     - int: the response status code
//     - *UsageType: the usage infomation of the specific uid
//     - error: the request error
func (c *Client) GetUsage(uid string, start, end *time.Time,
	showSummary, showEntries bool) (int, *UsageType, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(uid) != 0 {
		args.Add("uid", uid)
	}
	if start != nil {
		args.Add("start", start.Format(usageTimeFormat))
	}
	if end != nil {
		args.Add("end", end.Format(usageTimeFormat))
	}
	args.Add("show-entries", fmt.Sprintf("%v", showEntries))
	args.Add("show-summary", fmt.Sprintf("%v", showSummary))

	body, status, err := c.sendRequest("GET", "/usage", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}
	result := &UsageType{}
	if err := json.Unmarshal(body, result); err != nil {
		return status, nil, err
	}
	return status, result, nil
}

// DeleteUsage - delete the usage info of the radosgw service
//
// PARAMS:
//     - uid: user id string
//     - start: start timestamp
//     - end: end timestamp, not include
//     - deleteAll: delete all usage info, default is true
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) DeleteUsage(uid string, start, end *time.Time, deleteAll bool) (int, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(uid) != 0 {
		args.Add("uid", uid)
	}
	if start != nil {
		args.Add("start", start.Format(usageTimeFormat))
	}
	if end != nil {
		args.Add("end", end.Format(usageTimeFormat))
	}
	args.Add("deleteAll", fmt.Sprintf("%v", deleteAll))

	body, status, err := c.sendRequest("DELETE", "/usage", args, nil, nil)
	if err != nil {
		return status, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, fmt.Errorf("%s", string(body))
	}
	return status, nil
}
