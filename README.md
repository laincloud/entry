# Entry

A special LAIN application allowing developer entering his application's container to debug.

## Installation

- Download the codes, build and deploy in LAIN application workflow


- Grant access on `swarm.lain` to **entry**

    ```
    ETCD_AUTHORITY=127.0.0.1:4001 calicoctl profile entry rule show
    ```

    There may be some output like this

    ```
Inbound rules:
    1 allow from tag lain
    2 allow from tag entry
Outbound rules:
    1 deny tcp to ports 9003 cidr 192.168.77.0/24
    2 deny tcp to ports 9002 cidr 192.168.77.0/24
    3 deny tcp to ports 7001 cidr 192.168.77.0/24
    4 deny tcp to ports 4001 cidr 192.168.77.0/24
    5 deny tcp to ports 2376 cidr 192.168.77.0/24
    6 deny tcp to ports 2375 cidr 192.168.77.0/24
    7 allow
    ```

    Remove the `deny` rule for `swarm.lain` according to the port(This case is 2376).

    ```
    ETCD_AUTHORITY=127.0.0.1:4001 calicoctl profile entry rule remove outbound deny tcp to ports 2376 cidr 192.168.77.0/24
    ```

- Grant access on `lainlet.lain` to **entry**

    ```
    etcdctl set /lain/config/super_apps/entry {}
    ```

## License
Entry is released under [MIT](https://github.com/laincloud/entry/blob/master/LICENSE) license.
