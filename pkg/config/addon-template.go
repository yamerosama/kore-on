package config

const AddonTemplate = `

##################################################################################													                                      	##
## You can check supported applications with the 'list' command  	        			##
## and add applications to install. It can be used as below.				          	##
##                                                                              ##
## ※ If both "values" and "value_file" exist, "values" is used.				        	##
## -- Sample --													                                      	##
## [apps.application-name]													                          	##
## install = true																                                ##
## chart_ref_name = "xxx"												                            		##
## chart_ref = "<https://helm-chart-address or helm-package-address>"	  ##
## chart_name = "<Chart name to install>"                                       ##
## values="""																	                                  ##
## <Helm chart values input here>															                              ##
## """																	                                    		##
## value_file = "<helm chart values file path input>"								                  	##
##################################################################################

[addon]
## Required
## - k8s-master-ip: K8s control plane node ip address. (Deployment runs on this node.)
##					If you want to deploy locally, you must use the --kubeconfig option.
## - 
## Optional
## - ssh-port: K8s Controlplane Node ssh port (default: 22)
## - addon-data-dir: addon data(helm vales, k8s deployment yaml) dir (default: "/data/addon") 
## - closed-network: Enable Air Gap (default: false)
## - 
k8s-master-ip = "192.168.77.234"
#ssh-port = 22
addon-data-dir = "/data/addon"
close-network = false

[apps.csi-driver-nfs]
## Required
## - install: Choose to proceed with installation.
## - storage-ip: Storage node ip address.
## - volume-dir: Storage node data directory. (default: /data/storage)
## - nfs_version: Nfs-server version.
## - shared_volume_dir: shared directroy (default: /data/storage)
## -
install = true
chart_ref_name = "cube"
chart_ref = "https://hcapital-harbor.acloud.run/chartrepo/cube"
chart_name = "csi-driver-nfs"
values_file = "./csi-driver-nfs-values.yaml"
values = """
storageClass:
  create: true
  parameters:
    mountOptions:
    - nfsvers=4.1
    server: 192.168.77.239
    share: /data/storage
"""

[apps.koreboard]
## Required
## - install: Choose to proceed with installation.
## -
install = true
chart_ref_name = ""
chart_ref = "https://github.com/kore3lab/kore-dashboard/raw/master/scripts/install/kubernetes/kore-board-0.5.4.tgz"
chart_name = ""
values_file = "./koreboard-values.yaml"
values = """
"""

[apps.elasticsearch]
## Required
## - install: Choose to proceed with installation.
## - values: Input Helm Chart Values 
## -
install = true
chart_ref_name = "bitnami"
chart_ref = "https://charts.bitnami.com/bitnami"
chart_name = "elasticsearch"
values_file = ""
values = """
global:
  kibanaEnabled: true
  storageClass: "nfs-csi"
kibana:
  ingress:
    enabled: true
    hostname: kibana.test.com
    annotations:
      kubernetes.io/ingress.class: nginx
      cert-manager.io/cluster-issuer: letsencrypt-staging
    tls: true
"""

[apps.fluent-bit]
## Required
## - install: Choose to proceed with installation.
## -
install = true
chart_ref_name = "fluent"
chart_ref = "https://fluent.github.io/helm-charts"
chart_name = "fluent-bit"
values_file = ""
values = """
"""

`
