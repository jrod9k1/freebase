// Modified rip of this gist https://gist.github.com/unakatsuo/9a670058653a7aaaf2d83c032360647e
// This file is unused since we switches to the DNAT/REDIRECT method but keeping it in case we need it's functionality later

package main

import (
    "fmt"
    "net"
    "os"
    "syscall"

    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
)

func open(ifName string) (net.PacketConn, error) {
    fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
    if err != nil {
        return nil, fmt.Errorf("Failed open socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW): %s", err)
    }
    syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)

    if ifName != "" {
        //iface, err := net.InterfaceByName(ifName)
        _, err := net.InterfaceByName(ifName)
        if err != nil {
            return nil, fmt.Errorf("Failed to find interface: %s: %s", ifName, err)
        }
        syscall.SetsockoptString(fd, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, ifName)
    }

    conn, err := net.FilePacketConn(os.NewFile(uintptr(fd), fmt.Sprintf("fd %d", fd)))
    if err != nil {
        return nil, err
    }
    return conn, nil
}

func buildUDPPacket(dst, src *net.UDPAddr, payl []byte) ([]byte, error) {
    buffer := gopacket.NewSerializeBuffer()
    //payload := gopacket.Payload("HELLO")
    payload := gopacket.Payload(payl)
    ip := &layers.IPv4{
        DstIP:    dst.IP,
        SrcIP:    src.IP,
        Version:  4,
        TTL:      64,
        Protocol: layers.IPProtocolUDP,
    }
    udp := &layers.UDP{
        SrcPort: layers.UDPPort(src.Port),
        DstPort: layers.UDPPort(dst.Port),
    }
    if err := udp.SetNetworkLayerForChecksum(ip); err != nil {
        return nil, fmt.Errorf("Failed calc checksum: %s", err)
    }
    if err := gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}, ip, udp, payload); err != nil {
        return nil, fmt.Errorf("Failed serialize packet: %s", err)
    }
    return buffer.Bytes(), nil
}
