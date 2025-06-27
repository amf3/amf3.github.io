package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/go-plugins-helpers/volume"
)

const pluginRoot = "/var/lib/myplugin/volumes"

type myDriver struct{}

func (d *myDriver) Create(req *volume.CreateRequest) error {
	volPath := filepath.Join(pluginRoot, req.Name)
	log.Printf("Create volume: %s", volPath)
	return os.MkdirAll(volPath, 0755)
}

func (d *myDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	volPath := filepath.Join(pluginRoot, req.Name)

	// Write a hello.txt file
	helloFile := filepath.Join(volPath, "hello.txt")
	err := os.WriteFile(helloFile, []byte("Hello, world!\n"), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write hello.txt: %w", err)
	}

	log.Printf("Mount volume: %s -> %s", req.Name, volPath)
	return &volume.MountResponse{Mountpoint: volPath}, nil
}

func (d *myDriver) Unmount(req *volume.UnmountRequest) error {
	volPath := filepath.Join(pluginRoot, req.Name)
	helloFile := filepath.Join(volPath, "hello.txt")

	// Simulate cleanup
	if err := os.Remove(helloFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unmount cleanup error: %w", err)
	}

	log.Printf("Unmount volume: %s (removed hello.txt)", req.Name)
	return nil
}

func (d *myDriver) Remove(req *volume.RemoveRequest) error {
	volPath := filepath.Join(pluginRoot, req.Name)

	log.Printf("Remove volume: %s", volPath)
	err := os.RemoveAll(volPath)
	if err != nil {
		return fmt.Errorf("failed to remove volume: %w", err)
	}

	return nil
}

func (d *myDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	volPath := filepath.Join(pluginRoot, req.Name)

	// Confirm it exists
	info, err := os.Stat(volPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("volume %s not found", req.Name)
	} else if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid volume path: %w", err)
	}

	return &volume.GetResponse{
		Volume: &volume.Volume{
			Name:       req.Name,
			Mountpoint: volPath,
		},
	}, nil
}

func (d *myDriver) List() (*volume.ListResponse, error) {
	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	var volumes []*volume.Volume
	for _, entry := range entries {
		if entry.IsDir() {
			volPath := filepath.Join(pluginRoot, entry.Name())
			volumes = append(volumes, &volume.Volume{
				Name:       entry.Name(),
				Mountpoint: volPath,
			})
		}
	}

	return &volume.ListResponse{Volumes: volumes}, nil
}

func (d *myDriver) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "local", // or "global" for multi-host plugins
		},
	}
}

func (d *myDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	volPath := filepath.Join(pluginRoot, req.Name)
	return &volume.PathResponse{Mountpoint: volPath}, nil
}

func main() {
	driver := &myDriver{}
	h := volume.NewHandler(driver)
	log.Print("Starting myplugin ...")
	if err := h.ServeUnix("myplugin", 0); err != nil {
		log.Fatalf("plugin serve error: %v", err)
	}
}
