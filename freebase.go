package main

import (
    //"fmt"
    "log"
    "flag"
    "sync"
    "strconv"
    "net"
    "time"
)

var cache = make([]byte, 1400)
var cache_len int
var cache_mu sync.RWMutex // RW mutex here as we don't want readers stepping on each others toes but also don't want collisions during cache write c_c

var raw_conn net.PacketConn

func main() {
    source_ip := flag.String("ip", "", "IP of source server")
    source_port := flag.Int("port", 0, "Port of source server")
    bind_port := flag.Int("bport", 0, "Interception/reflection port to bind to on router")
    refresh_time := flag.Int("refresh", 60, "Cache time to live in seconds (default 60)")
    flag.Parse()

    log.Print("Freebase started!")
    log.Print("Params: source_ip:", *source_ip, " source_port:", *source_port, " bind_port:", *bind_port, " refresh_time:", *refresh_time)

    // TODO: Some validation / sanity checking of passed values?

    var err error
    raw_conn, err = open("eth0")
    if err != nil {
        log.Fatal(err)
    }

    var wg sync.WaitGroup
    wg.Add(2)

    go a2s_fetch(*refresh_time, *source_ip, *source_port)
    go a2s_serve(*bind_port, *source_ip, *source_port)

    wg.Wait()
}

// I split this out into 2 functions as I need to to fetch A2S immediately
// on startup to avoid having an empty cache before moving on to the tick
// loop. Maybe a less shit way to do this
func a2s_fetch(refresh_interval int, ip string, port int) {
    log.Print("Fetch thread spawned")

    do_a2s_fetch(ip, port)

    ttl := time.Duration(refresh_interval) * time.Second
    for range time.Tick(ttl) {
        do_a2s_fetch(ip, port)
    }
}

func do_a2s_fetch(ip string, port int) {
    log.Print("REMOVEME a2s tick")
    pkt, length, err := query_server(ip, port, 5) // TODO: maybe don't make timeout hardcoded later
    if err != nil {
        log.Print("Query server failed!")
        log.Print(err)
    }

    log.Print("REMOVEME server responded w: ", string(pkt))

    cache_mu.Lock()
    cache = pkt
    cache_len = length
    cache_mu.Unlock()
}

func a2s_serve(bind_port int, source_ip string, source_port int) {
    log.Print("Serve thread spawned")

    pc, err := net.ListenPacket("udp", ":" + strconv.Itoa(bind_port))
    if err != nil {
        log.Print("Could not bind to bport!")
        log.Fatal(err)
    }

    defer pc.Close()

    for {
        received_pkt := make([]byte, 100)
        n, addr, err := pc.ReadFrom(received_pkt)
        if err != nil {
            log.Print("REMOVEME fucked packet from ", addr)
            continue
        }

        go send_cache(pc, addr, received_pkt[:n], n, source_ip, source_port)
    }
}

func query_server(address string, port int, timeout int) ([]byte, int, error) {
    log.Print("REMOVEME query_server run")

    service := address + ":" + strconv.Itoa(port)
    raddr, err := net.ResolveUDPAddr("udp", service)

    conn, err := net.DialUDP("udp", nil, raddr)
    if err != nil {
        log.Print("Could not connect to source server!")
        return nil, 0, err
    }

    timeout_duration := time.Duration(timeout) * time.Second
    err = conn.SetReadDeadline(time.Now().Add(timeout_duration))
    if err != nil {
        log.Print("Timed out waiting for source server reply")
        return nil, 0, err
    }

    a2s_message := []byte("\xFF\xFF\xFF\xFFTSource Engine Query\x00")

    _, err = conn.Write(a2s_message)
    if err != nil {
        log.Print("Could not send A2S_INFO to source server!")
        return nil, 0, err
    }

    reply := make([]byte, 1400)
    n, _, err := conn.ReadFromUDP(reply)
    if err != nil {
        log.Print("Error reading source server reply!")
        return nil, 0, err
    }

    return reply, n, nil
}

// aye bruh pass the pointer
/*
func send_cache(pc net.PacketConn, addr net.Addr, pkt []byte, n int, source_ip string, source_port int) {
    log.Print("REMOVE ME sending pkt back to ", addr)

    source := &net.UDPAddr { // TODO: fix this garbage
        //IP: net.ParseIP(source_ip),
        //IP: net.ParseIP("142.202.137.10"),
        IP: net.ParseIP("142.202.137.10"),
        //Port: source_port,
        Port: 27016,
        //Port: 27016,
    }
    log.Print("YEET IP ", source_ip)
    log.Print("YEET PORT ", strconv.Itoa(source_port))

    dest := addr.(*net.UDPAddr)

    // TODO: RWLock here
    //cache_mu.RLock()
    b, err := buildUDPPacket(dest, source, cache[:cache_len])
    //cache_mu.RUnlock()
    if err != nil {
        log.Print("Failed building raw UDP packet")
    }

    _, err = raw_conn.WriteTo(b, &net.IPAddr{IP: dest.IP})
    if err != nil {
        log.Print("Failed sending cache to ", addr)
    }
}
*/

func send_cache(pc net.PacketConn, addr net.Addr, pkt []byte, n int, source_ip string, source_port int) {
    log.Print("REMOVE ME sending pkt back to ", addr)

    err := pc.SetWriteDeadline(time.Now().Add(time.Duration(5) * time.Second))
    if err != nil {
        log.Fatal(err)
    }

    n, err = pc.WriteTo(cache[:cache_len], addr)
    if err != nil{
        log.Fatal(err)
    }
}
