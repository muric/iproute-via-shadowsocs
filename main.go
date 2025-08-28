package main

import (
    "bufio"
    "encoding/json"
    "log"
    "io/ioutil"
    "net"
    "fmt"
    "os"
    "strings"
    "sync"
    "path/filepath"
    "github.com/vishvananda/netlink"
)

type Config struct {
    Gateway   string
    Interface string
}

func readConfig(filename string) (Config, error) {
    file, err := os.Open(filename)
    if err != nil {
        return Config{}, err
    }
    defer file.Close()

    config := Config{}
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue // Ignore invalid lines
        }
        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])

        switch key {
        case "gateway":
            config.Gateway = value
        case "interface":
            config.Interface = value
        }
    }

    if err := scanner.Err(); err != nil {
        return Config{}, err
    }

    return config, nil
}

func addRoute(destination, gateway, ifaceName string) error {
    iface, err := netlink.LinkByName(ifaceName)
    if err != nil {
        return fmt.Errorf("error reading interface %s: %v", ifaceName, err)
    }

    // Parse CIDR or single IP
    ip, ipNet, err := net.ParseCIDR(destination)
    if err != nil {
        ip = net.ParseIP(destination)
        if ip == nil {
            return fmt.Errorf("error parsing destination address %s: %v", destination, err)
        }
        ipNet = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)} // Mask for single IP
    }

    route := &netlink.Route{
        Dst:       ipNet,
        Gw:        net.ParseIP(gateway),
        LinkIndex: iface.Attrs().Index,
    }

    if err := netlink.RouteAdd(route); err != nil {
        return fmt.Errorf("error adding route %s via %s: %v", destination, gateway, err)
    }

    return nil
}

func main() {
    config, err := readConfig("iproute.conf")
    if err != nil {
        log.Printf("Error reading configuration: %v\n", err)
        return
    }

    dir := "data"

    var jsonFiles []string

    err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() && filepath.Ext(path) == ".json" {
            jsonFiles = append(jsonFiles, info.Name())
        }
        return nil
    })

    if err != nil {
        log.Printf("Error reading folder %s: %v\n", dir, err)
        return
    }

    for _, fileName := range jsonFiles {
        log.Println("working on: " + fileName )
        data, err := ioutil.ReadFile(filepath.Join(dir, fileName))
        if err != nil {
            log.Printf("Error reading file: %v\n", err)
            return
        }

        var destinations []string
        if err := json.Unmarshal(data, &destinations); err != nil {
            log.Printf("Error parsing JSON: %v\n", err)
            return
        }

        var wg sync.WaitGroup
        sem := make(chan struct{}, 100)

        for _, destination := range destinations {
            wg.Add(1)
            sem <- struct{}{}
            go func(dest string) {
                defer wg.Done()
                defer func() { <-sem }()

                if err := addRoute(dest, config.Gateway, config.Interface); err != nil {
			log.Printf("Error adding route for %s: %v\n", dest, err)
                }
            }(destination)
        }

        wg.Wait()
    }
}
