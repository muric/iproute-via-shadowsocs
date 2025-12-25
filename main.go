package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "strconv"

    "github.com/vishvananda/netlink"
)

type Config struct {
    Gateway          string
    Interface        string
    DefaultGateway   string
    DefaultInterface string
    GoroutineCount   int
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
            continue
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])

        switch key {
        case "gateway":
            config.Gateway = value
        case "interface":
            config.Interface = value
        case "default_gw":
            config.DefaultGateway = value
        case "default_interface":
            config.DefaultInterface = value
	case "goroutine_count":
	    config.GoroutineCount, err = strconv.Atoi(value)
	    if err != nil {
        	panic(err)
		}
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

    ip, ipNet, err := net.ParseCIDR(destination)
    if err != nil {
        ip = net.ParseIP(destination)
        if ip == nil {
            return fmt.Errorf("error parsing destination %s: %v", destination, err)
        }
        ipNet = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
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

func addRoutesFromDir(dir, gateway, iface string, gouroutinecount int) error {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        log.Printf("Directory %s does not exist — skipping\n", dir)
        return nil
    }

    var jsonFiles []string

    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && filepath.Ext(path) == ".json" {
            jsonFiles = append(jsonFiles, info.Name())
        }
        return nil
    })
    if err != nil {
        return fmt.Errorf("error reading folder %s: %v", dir, err)
    }

    if len(jsonFiles) == 0 {
        log.Printf("No route files found in %s — skipping\n", dir)
        return nil
    }

    for _, fileName := range jsonFiles {
        log.Println("Processing:", fileName)
        data, err := ioutil.ReadFile(filepath.Join(dir, fileName))
        if err != nil {
            log.Printf("\033[31mError reading file %s: %v\033[0m\n", fileName, err)
            continue
        }

        var destinations []string
        if err := json.Unmarshal(data, &destinations); err != nil {
            log.Printf("\033[31mError parsing JSON %s: %v\033[0m\n", fileName, err)
            continue
        }

        var wg sync.WaitGroup
        sem := make(chan struct{}, gouroutinecount)

        for _, dest := range destinations {
            wg.Add(1)
            sem <- struct{}{}
            go func(d string) {
                defer wg.Done()
                defer func() { <-sem }()
                if err := addRoute(d, gateway, iface); err != nil {
                    log.Printf("\033[31mError adding route for %s via %s dev %s: %v\033[0m\n", d, gateway, iface, err)
                }
            }(dest)
        }

        wg.Wait()
    }
    return nil
}

func main() {
    config, err := readConfig("iproute.conf")
    if err != nil {
        log.Fatalf("\033[31mError reading configuration: %v\033[0m", err)
    }

    mainDir := "data"
    defaultDir := "default_route"

    if config.Interface != "" && config.Gateway != "" {
        log.Println("Adding routes for interface:", config.Interface)
        if err := addRoutesFromDir(mainDir, config.Gateway, config.Interface, config.GoroutineCount); err != nil {
            log.Printf("\033[31mError adding routes: %v\033[0m\n", err)
        }
    }

    if config.DefaultInterface != "" && config.DefaultGateway != "" {
        log.Println("Adding routes for default interface:", config.DefaultInterface)
        if err := addRoutesFromDir(defaultDir, config.DefaultGateway, config.DefaultInterface, config.GoroutineCount); err != nil {
            log.Printf("\033[31mError adding default routes: %v\033[0m\n", err)
        }
    }
}

