package config

const Template = `

[koreon]
## Required
## - local-repository: local repository 서비스 url (Required when selecting the closed network.)
## - local-repository-archive-file: local repository packages archive file path (Required when selecting the closed network.)
## Optional
## - cluster-name: use cluster name in config context (default: "kubernetes")
## - install-dir: installation scripts(harbor, shell scripts) save directory (default: "/var/lib/kore-on")
## - cert-validity-days: SSL validity days(default: 36500)
## - debug-mode: verbose 옵션 사용 여부 선택 (default: false)
## - closed-network: Air Gap 선언 (default: false)
#cluster-name = "test-cluster"
#install-dir = "/var/lib/kore-on"
#cert-validity-days = 36500
#debug-mode = true
#closed-network = false
#local-repository = "http://192.168.77.239:8080"
#local-repository-archive-file = "/tmp/koreon/local-repo.20220224_071700.tgz"

[kubernetes]
## Required
## - 
## Optional
## - version: kubernetes version (default: "latest")
## - container-runtime: use k8s cri (only containerd)
## - kube-proxy-mode: use k8s proxy mode [iptables | ipvs] (default: "ipvs")
## - service-cidr: k8s service network cidr (default: "10.96.0.0/20")
## - pod-cidr: k8s pod network cidr (default: "10.4.0.0/24")
## - node-port-range: k8s node port network range (default: "30000-32767")
## - audit-log-enable: k8s audit log enabled (default: true)
## - api-sans: k8s apiserver SAN 추가 [--apiserver-cert-extra-sans 설정과 동일] (default: master[0] ip address)
version = "v1.23.12"
#container-runtime = "containerd"
#kube-proxy-mode = "ipvs"
#service-cidr ="172.20.0.0/24"
#pod-cidr="10.10.0.0/24"
#node-port-range="30000-32767"
#api-sans = ["192.168.77.234"]

[kubernetes.etcd]
## Required
## - ip: k8s control plane nodes ip address. (this is using it to generate an inventory)
## - private-ip: K8s control plane nodes private ip address. (this is using it to generate an inventory)
##               If you use the same IP address, you can skip it.
## Optional
## - external-etcd: used external etcd than input the ip and private-ip address (default: false)
##                  not used than skip ip address. it is used control plane nodes as automatic.
#external-etcd = true
#ip = ["x.x.x.x"]
#private-ip = ["x.x.x.x"]


[kubernetes.calico]
## Required
## - 
## Optional
## - version: calico version (default: "latest")
## - vxlan-mode: calico VXLAN mode activate (default: false)
version = "v3.24"
#vxlan-mode = true


[node-pool]
## Required
## - 
## Optional
## - data-dir: data(backup, docker, log, kubelet, etcd, k8s-audit, containerd) root dir (default: "/data") 
data-dir = "/data"

[node-pool.security]
## Required
## - ssh-user-id: node user id (You can skip this entry by using the --user command option)
## - private-key-path: ssh private key path (You can skip this entry by using the --private-key-path command option)
#ssh-user-id = "cloud-user"
#private-key-path = "/Users/dongmook/DEV_WORKS/cert_ssh/rhel/cloud-user"

[node-pool.master]
## Required
## - ip: k8s control plane nodes ip address. (this is using it to generate an inventory)
## - private-ip: K8s control plane nodes private ip address. (this is using it to generate an inventory)
##               If you use the same IP address, you can skip it.
## Optional
## - lb-ip: loadbalancer ip address (default: master[0] node ip address)
## - isolated: K8s control plane nodes isolated (default: true)
## - haproxy-install: used internal load-balancer (default: true)
## - lb-ip: Enter the IP address when using a load balancer (default: master[0] ip address)
## - lb-port: Enter the port when using a load balancer (default: "6443")
ip = ["192.168.77.234","192.168.77.235","192.168.77.236"]
private-ip = ["3.0.0.1","3.0.0.2","3.0.0.3"]
#isolated = true
#haproxy-install = true
#lb-ip = "192.168.77.234"
#lb-port = "6443"

[node-pool.node]
## Required
## - ip: k8s work nodes ip address. (this is using it to generate an inventory)
## - private-ip: K8s work nodes private ip address. (this is using it to generate an inventory)
##               If you use the same IP address, you can skip it.
## Optional
ip = ["192.168.77.237", "192.168.77.238"]
private-ip = ["3.0.0.11", "3.0.0.12"]

[private-registry]
## Required
## - registry-version: private registry nodes ip address. This is a required entry used when installing a private registry.
##                (this is using it to generate an extra vars)
## - registry-ip: private registry nodes ip address. This is a required entry used when installing a private registry.
##                (this is using it to generate an inventory)
## - private-ip: K8s work node private ip address. This is a required entry used when installing a private registry.
##               If you use the same IP address, you can skip it.
##               (this is using it to generate an inventory)
## - registry-domain: K8s registry configuration (this is using it to generate an extra vars)
## Optional
## - install: private registry install (default: false)
## - data-dir: private registry data directory (default: "/data/harbor")
## - registry-archive-file: registry archive file path (default: "")
## - public-cert: public cert activate (default: false)
install = true
registry-version = "v2.6.0"
registry-ip = "192.168.77.239"
private-ip = "3.0.0.50"
registry-domain = "192.168.77.239"
data-dir = "/data/harbor"
#registry-archive-file = "/tmp/koreon/harbor.20220224_072307.tgz"
public-cert = false

[private-registry.cert-file]
## Required
## - ssl-certificate: The certificate path used when using public-cert.
##                    This is a required field used when using a public certificate.
## - ssl-certificate-key: The certificate-key used when using public-cert.
##                        This is a required field used when using a public certificate.
## Optional
#ssl-certificate = ""
#ssl-certificate-key = ""

[shared-storage]
## Required
## - storage-ip: Storage node ip address.
##               This is a required field used when installing the nfs server.
##               (this is using it to generate an inventory and generate an extra vars)
## - private-ip: Storage node ip address.
##               This is a required field used when installing the nfs server.
##               If you use the same IP address, you can skip it.
##               (this is using it to generate an inventory)
## - volume-dir: Storage node data directory. (defalue: /storage)
##               This is a required field used when installing the nfs server.
##               (this is using it to generate an extra vars)
## Optional
## - install: NFS Server Installation (default: false)
install = true
storage-ip = "192.168.77.239"
private-ip = "3.0.0.50"
volume-dir = "/data/storage"
`
