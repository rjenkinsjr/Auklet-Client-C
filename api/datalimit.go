package api

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// CellularConfig defines a limit and date for devices that use a cellular
// connection.
type CellularConfig struct {
	// Limit is the maximum number of application layer
	// megabytes/period that the client may send over a
	// cellular connection.
	Limit *int `json:"cellular_data_limit"`

	// Date is the day of the month that delimits a cellular
	// data plan period. Valid values are within [1, 28].
	Date int `json:"normalized_cell_plan_date"`
}

// DataLimit represents parameters that control the client's use of data.
type DataLimit struct {
	Config struct {
		// EmissionPeriod is the time in seconds the client is to wait
		// between emission requests to the agent.
		EmissionPeriod int `json:"emission_period"`
		Storage        struct {
			// Limit is the maximum number of megabytes the client
			// may use to store unsent messages. If nil, there is no
			// storage limit.
			Limit *int `json:"storage_limit"`
		} `json:"storage"`
		Cellular CellularConfig `json:"data"`
	} `json:"config"`
}

// GetDataLimit returns a DataLimit from the dataLimit endpoint.
func GetDataLimit(appID string) (l DataLimit) {
	resp := get(fmt.Sprintf(dataLimitEP, appID), "application/json")
	if resp == nil {
		return
	}
	if resp.StatusCode != 200 {
		errorlog.Printf("api.DataLimit: unexpected status %v", resp.Status)
	}
	d := json.NewDecoder(resp.Body)
	err := d.Decode(&l)
	if err != nil && err != io.EOF {
		errorlog.Print(err)
	}
	return
}
