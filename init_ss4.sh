ip link add name my_bridge type bridge
ip addr add 192.168.11.100/24 dev my_bridge
ip link set my_bridge up
sudo ip link set enp0s31f6 master my_bridge
