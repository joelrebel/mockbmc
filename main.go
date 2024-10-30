package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	// DefaultPort to listen on. Seen in the bmclib docs.
	DefaultPort = 8800

	// StateOn is the power "on" state.
	StateOn = "on"

	// StateOff is the power "off" state.
	StateOff = "off"
)

// MockBMC is a simple mocked BMC device.
type MockBMC struct {
	// Addr to listen on. Eg: :8800 for example.
	Addr string

	// State of the device power. This gets read and changed by the API.
	State string
}

func mustReadFile(filename string) []byte {
	fixture := "fixtures" + "/" + filename
	fh, err := os.Open(fixture)
	if err != nil {
		log.Fatal(err)
	}

	defer fh.Close()

	b, err := io.ReadAll(fh)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func (obj *MockBMC) endpointFunc(file, method string, retStatus int, retHeader map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method &&
			r.Method != http.MethodPatch { // purge check on patch method if set pxe request is attempted
			resp := fmt.Sprintf("unexpected request - url: %s, method: %s", r.URL, r.Method)
			_, _ = w.Write([]byte(resp))
		}

		for k, v := range retHeader {
			w.Header().Add(k, v)
		}

		w.WriteHeader(retStatus)
		if file != "" {
			_, _ = w.Write(mustReadFile(file))
		}

		return
	}
}

// Run kicks this all off.
func (obj *MockBMC) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", obj.endpointFunc("service_root.json", http.MethodGet, 200, nil))
	// login
	sessionHeader := map[string]string{
		"X-Auth-Token": "t5tpiajo89fyvvel5434h9p2l3j69hzx",
		"Location":     "/redfish/v1/SessionService/Sessions/1",
	}
	mux.HandleFunc("/redfish/v1/SessionService/Sessions", obj.endpointFunc("session_service.json", http.MethodPost, 201, sessionHeader))
	mux.HandleFunc("/redfish/v1/Systems", obj.endpointFunc("systems.json", http.MethodGet, 200, nil))
	mux.HandleFunc("/redfish/v1/Systems/1", obj.endpointFunc("systems_1.json", http.MethodGet, 200, nil))
	// set pxe - we can't have two routes with the same pattern
	// mux.HandleFunc("/redfish/v1/Systems/1", obj.endpointFunc("", http.MethodPatch, 200, nil))
	// power on/off
	mux.HandleFunc("/redfish/v1/Systems/1/Actions/ComputerSystem.Reset", obj.endpointFunc("", http.MethodPost, 200, nil))
	// logoff
	mux.HandleFunc("/redfish/v1/SessionService/Sessions/1", obj.endpointFunc("session_delete.json", http.MethodDelete, 200, nil))

	return http.ListenAndServeTLS(obj.Addr, "snakeoil/localhost.crt", "snakeoil/localhost.key", mux)
}

// Main program that returns error.
func Main() error {
	mock := &MockBMC{
		Addr:  fmt.Sprintf("localhost:%d", DefaultPort),
		State: "off",
	}
	return mock.Run()
}

func main() {
	fmt.Printf("main: %+v\n", Main())
}
