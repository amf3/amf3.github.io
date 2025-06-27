package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const pluginRoot = "/var/lib/myplugin/volumes"

type VolumeRequest struct {
	Name string `json:"Name"`
}

type VolumeResponse struct {
	Mountpoint string `json:"Mountpoint,omitempty"`
	Err        string `json:"Err,omitempty"`
}

func createVolumeHandler(w http.ResponseWriter, r *http.Request) {
	var req VolumeRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	log.Printf("Create volume: %s\n", req.Name)
	writeJSON(w, VolumeResponse{})
}

func mountVolumeHandler(w http.ResponseWriter, r *http.Request) {
	var req VolumeRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	volPath := filepath.Join(pluginRoot, req.Name)
	if err := os.MkdirAll(volPath, 0755); err != nil {
		writeJSON(w, VolumeResponse{Err: err.Error()})
		return
	}

	helloFile := filepath.Join(volPath, "hello.txt")
	err := os.WriteFile(helloFile, []byte("Hello, world!\n"), 0644)
	if err != nil {
		writeJSON(w, VolumeResponse{Err: err.Error()})
		return
	}

	log.Printf("Mounted volume %s at %s\n", req.Name, volPath)
	writeJSON(w, VolumeResponse{Mountpoint: volPath})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func main() {
	http.HandleFunc("/VolumeDriver.Create", createVolumeHandler)
	http.HandleFunc("/VolumeDriver.Mount", mountVolumeHandler)

	socket := "/run/docker/plugins/myplugin.sock"
	_ = os.Remove(socket) // Ensure clean start

	log.Printf("Starting plugin on unix socket: %s\n", socket)
	err := http.ListenAndServe("unix://"+socket, nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
