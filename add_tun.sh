#!/usr/bin/env bash
ip tuntap add dev tun2 mode tun
ip a add 10.0.0.1/24 dev tun2
ip link set dev tun2 up
