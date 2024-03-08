ip link add name my_bridge type bridge
ip addr add 192.168.11.100/16 dev my_bridge
ip addr add 192.168.11.101/16 dev my_bridge
ip link set my_bridge up