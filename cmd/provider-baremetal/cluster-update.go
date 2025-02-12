package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kore-on/cmd/koreonctl/conf"
	"kore-on/cmd/koreonctl/conf/templates"
	"kore-on/pkg/cluster/kubemethod"
	"kore-on/pkg/logger"
	"kore-on/pkg/model"
	"kore-on/pkg/model/k8s"
	"kore-on/pkg/utils"

	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/apenella/go-ansible/pkg/execute"
	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/apenella/go-ansible/pkg/stdoutcallback/results"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// Commands structure
type strClusterUpdateCmd struct {
	dryRun        bool
	verbose       bool
	inventory     string
	tags          string
	playbookFiles []string
	privateKey    string
	user          string
	command       string
	kubeconfig    string
	extravars     map[string]interface{}
}

var err error

func ClusterUpdateCmd() *cobra.Command {
	clusterUpdate := &strClusterUpdateCmd{}

	cmd := &cobra.Command{
		Use:          "update [flags]",
		Short:        "Update kubernetes cluster(node scale in/out)",
		Long:         "This command update the Kubernetes cluster nodes (node scale in/out)",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterUpdate.run()
		},
	}

	// SubCommand add
	cmd.AddCommand(
		GetKubeConfigCmd(),
		UpdateInitCmd(),
	)

	// SubCommand validation
	utils.CheckCommand(cmd)

	// Default value for command struct
	clusterUpdate.command = "update"
	clusterUpdate.tags = ""
	clusterUpdate.inventory = "./internal/playbooks/koreon-playbook/inventory/inventory.ini"
	clusterUpdate.playbookFiles = []string{
		"./internal/playbooks/koreon-playbook/cluster-update.yaml",
	}

	f := cmd.Flags()
	f.BoolVar(&clusterUpdate.verbose, "verbose", false, "verbose")
	f.BoolVarP(&clusterUpdate.dryRun, "dry-run", "d", false, "dryRun")
	f.StringVar(&clusterUpdate.tags, "tags", clusterUpdate.tags, "Ansible options tags")
	f.StringVarP(&clusterUpdate.privateKey, "private-key", "p", "", "Specify ssh key path")
	f.StringVarP(&clusterUpdate.user, "user", "u", "", "login user")
	f.StringVar(&clusterUpdate.kubeconfig, "kubeconfig", "", "get kubeconfig")

	return cmd
}

func GetKubeConfigCmd() *cobra.Command {
	getKubeConfig := &strClusterUpdateCmd{}

	cmd := &cobra.Command{
		Use:          "get-kubeconfig [flags]",
		Short:        "Get Kubeconfig file",
		Long:         "This command get kubeconfig file in k8s controlplane node.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getKubeConfig.run()
		},
	}

	getKubeConfig.command = "get-kubeconfig"
	getKubeConfig.tags = ""
	getKubeConfig.inventory = "./internal/playbooks/koreon-playbook/inventory/inventory.ini"
	getKubeConfig.playbookFiles = []string{
		"./internal/playbooks/koreon-playbook/cluster-get.yaml",
	}

	f := cmd.Flags()
	f.BoolVarP(&getKubeConfig.verbose, "verbose", "v", false, "verbose")
	f.BoolVarP(&getKubeConfig.dryRun, "dry-run", "d", false, "dryRun")
	f.StringVar(&getKubeConfig.tags, "tags", getKubeConfig.tags, "Ansible options tags")
	f.StringVarP(&getKubeConfig.privateKey, "private-key", "p", "", "Specify ssh key path")
	f.StringVarP(&getKubeConfig.user, "user", "u", "", "login user")

	return cmd
}

func UpdateInitCmd() *cobra.Command {
	updateInit := &strClusterUpdateCmd{}

	cmd := &cobra.Command{
		Use:          "init [flags]",
		Short:        "Get Installed Config file",
		Long:         "This command get installed config file in k8s controlplane node.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateInit.run()
		},
	}

	updateInit.command = "update-init"
	updateInit.tags = ""
	updateInit.inventory = "./internal/playbooks/koreon-playbook/inventory/inventory.ini"
	updateInit.playbookFiles = []string{
		"./internal/playbooks/koreon-playbook/cluster-get.yaml",
	}

	f := cmd.Flags()
	f.BoolVarP(&updateInit.verbose, "verbose", "v", false, "verbose")
	f.BoolVarP(&updateInit.dryRun, "dry-run", "d", false, "dryRun")
	f.StringVar(&updateInit.tags, "tags", updateInit.tags, "Ansible options tags")
	f.StringVarP(&updateInit.privateKey, "private-key", "p", "", "Specify ssh key path")
	f.StringVarP(&updateInit.user, "user", "u", "", "login user")
	f.StringVar(&updateInit.kubeconfig, "kubeconfig", "", "get kubeconfig")

	return cmd
}

func (c *strClusterUpdateCmd) run() error {
	koreOnConfigFileName := conf.KoreOnConfigFile
	koreOnConfigFilePath := utils.IskoreOnConfigFilePath(koreOnConfigFileName)
	koreonToml, errBool := utils.ValidateKoreonTomlConfig(koreOnConfigFilePath, "cluster-update")
	if !errBool {
		message := "Settings are incorrect. Please check the 'korean.toml' file!!"
		logger.Fatal(fmt.Errorf("%s", message))
	}

	// koreonToml Default value
	koreonToml.KoreOn.FileName = koreOnConfigFileName

	// current pocessing directory
	dir, err := utils.Dirname("../..")
	if err != nil {
		logger.Fatal(err)
	}
	if dir == "/build" {
		dir = ""
	}
	koreonToml.KoreOn.WorkDir = dir + "/" + conf.KoreOnConfigFileSubDir

	koreonToml.KoreOn.CommandMode = c.command
	if c.command == "get-kubeconfig" {
		koreonToml.Kubernetes.GetKubeConfig = true
	}

	if len(c.playbookFiles) < 1 {
		return fmt.Errorf("[ERROR]: %s", "To run ansible-playbook playbook file path must be specified")
	}

	if len(c.inventory) < 1 {
		return fmt.Errorf("[ERROR]: %s", "To run ansible-playbook an inventory must be specified")
	}

	if len(c.privateKey) < 1 {
		return fmt.Errorf("[ERROR]: %s", "To run ansible-playbook an privateKey must be specified")
	}

	if len(c.user) < 1 {
		return fmt.Errorf("[ERROR]: %s", "To run ansible-playbook an ssh login user must be specified")
	}

	if c.command != "update-init" && c.command != "get-kubeconfig" && len(c.kubeconfig) < 1 {
		return fmt.Errorf("[ERROR]: %s", "To run this ansible-playbook an kubeconfig option must be specified.\n You can get kubeconfig with 'get-kubeconfig' command")
	}

	var master []k8s.Node
	var node []k8s.Node
	updateNodeIP := make(map[int]string)
	updateNodePrivateIP := make(map[int]string)
	updateNodeName := make(map[int]string)
	updateType := ""

	if len(koreonToml.NodePool.Node.PrivateIP) == 0 {
		koreonToml.NodePool.Node.PrivateIP = koreonToml.NodePool.Node.IP
	}

	if c.command == "update" {
		// Get k8s clientset
		kubeconfigPath, _ := filepath.Abs(c.kubeconfig)
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			logger.Fatal(err)
		}
		var lbIP []string

		if koreonToml.NodePool.Master.LbIP != "" {
			lbIP[0] = koreonToml.NodePool.Master.LbIP
		} else {
			lbIP = koreonToml.NodePool.Master.IP
		}

		chekServerAddress := checkKubeconfig(lbIP, config.Host)

		if !chekServerAddress {
			logger.Fatal("The cluster is unreachable. Check the kubeconfig server address.")
		}

		client, err := kubemethod.CreateK8sClient(config)
		if err != nil {
			logger.Fatal(err)
		}

		// Get K8s Cluster Nodes
		kubeNodes, err := kubemethod.GetNodeList(client)
		if err != nil {
			logger.Fatal(err)
		}

		for _, v := range kubeNodes {
			if strings.Contains(v.Role, "control-plane") {
				master = append(master, v)
			} else {
				node = append(node, v)
			}
		}

		updateNodePrivateIP = getNodeIP(koreonToml.NodePool.Node.PrivateIP)
		updateNodeIP = getNodeIP(koreonToml.NodePool.Node.IP)
		// koreon.toml PrivateIP 값이 있을때

		if len(koreonToml.NodePool.Node.PrivateIP) == len(node) {
			var cnt int
			for _, v := range koreonToml.NodePool.Node.PrivateIP {
				for _, j := range node {
					if j.InternalIP == v {
						cnt = cnt + 1
					}
				}
			}
			if cnt == len(node) {
				logger.Fatal("Same as the current cluster node list. There are no node entries to update. Please check node pool input.")
			}
		}
		if len(koreonToml.NodePool.Node.PrivateIP) > 0 {
			// Node 추가
			if len(koreonToml.NodePool.Node.PrivateIP) > len(node) {
				updateType = "ADD"
				for k, v := range koreonToml.NodePool.Node.PrivateIP {
					for _, j := range node {
						if strings.Contains(j.InternalIP, v) {
							delete(updateNodePrivateIP, k)
							delete(updateNodeIP, k)
						}
					}
				}
			}
			// Node 삭제
			if len(koreonToml.NodePool.Node.PrivateIP) < len(node) {
				updateType = "DELETE"
				c.playbookFiles = []string{
					"./internal/playbooks/koreon-playbook/cluster-remove-node.yaml",
				}
				for k, v := range node {
					updateNodePrivateIP[k] = v.InternalIP
					updateNodeIP[k] = v.AnsibleSshHost
					updateNodeName[k] = v.Name
				}
				for k, v := range node {
					for _, j := range koreonToml.NodePool.Node.PrivateIP {
						if strings.Contains(j, v.InternalIP) {
							delete(updateNodePrivateIP, k)
							delete(updateNodeIP, k)
						}
					}
				}
			}
		}
		// koreon.toml PrivateIP 값이 없을때
		if len(koreonToml.NodePool.Node.PrivateIP) == 0 {
			// Node 추가
			if len(koreonToml.NodePool.Node.PrivateIP) > len(node) {
				updateType = "ADD"
				for k, v := range koreonToml.NodePool.Node.IP {
					for _, j := range node {
						if strings.Contains(j.InternalIP, v) {
							delete(updateNodeIP, k)
						}
					}
				}
			}
			//Node 삭제
			if len(koreonToml.NodePool.Node.PrivateIP) < len(node) {
				updateType = "DELETE"
				c.playbookFiles = []string{
					"./internal/playbooks/koreon-playbook/cluster-remove-node.yaml",
				}
				for k, v := range node {
					updateNodePrivateIP[k] = v.InternalIP
					updateNodeIP[k] = v.AnsibleSshHost
					updateNodeName[k] = v.Name
				}
				for k, v := range node {
					for _, j := range koreonToml.NodePool.Node.IP {
						if strings.Contains(j, v.InternalIP) {
							delete(updateNodeIP, k)
						}
					}
				}
			}
		}
		if len(updateNodeIP) == 0 {
			logger.Fatal("There are no node entries to update. Please check node pool input.")
		}
	}

	// Make provision data
	data := model.KoreonctlText{}
	data.KoreOnTemp = koreonToml
	data.Master = master
	data.Node = node
	for k, v := range updateNodeIP {
		data.UpdateNode.IP = append(data.UpdateNode.IP, v)
		data.UpdateNode.PrivateIP = append(data.UpdateNode.PrivateIP, updateNodePrivateIP[k])
		data.UpdateNode.Name = append(data.UpdateNode.Name, updateNodeName[k])
	}
	koreonToml.NodePool.Node.Name = data.UpdateNode.Name
	koreonToml.NodePool.Node.IP = data.UpdateNode.IP
	koreonToml.NodePool.Node.PrivateIP = data.UpdateNode.PrivateIP
	koreonToml.KoreOn.Update = true

	// Processing template
	koreonctlText := template.New("ClusterUpdateText")

	// template func
	koreonctlText.Funcs(template.FuncMap(map[string]interface{}{
		"maxLength": func(item interface{}) map[string]int {
			maxlen := make(map[string]int)
			blankCnt := 2
			v := reflect.ValueOf(item)

			switch v.Kind() {
			case reflect.Invalid:
				return maxlen
			case reflect.Slice:
				for i := 0; i < v.Len(); i++ {
					r := v.Index(i)
					switch r.Kind() {
					case reflect.Struct:
						checkEmpty := i
						t := r.Type()
						for i := 0; i < r.NumField(); i++ {
							fieldValue := r.Field(i)
							field := t.Field(i)
							if checkEmpty == 0 {
								maxlen[field.Name] = len(fmt.Sprintf("%s", fieldValue.Interface())) + blankCnt
							} else {
								if maxlen[field.Name] < len(fmt.Sprintf("%s", fieldValue.Interface()))+blankCnt {
									maxlen[field.Name] = len(fmt.Sprintf("%s", fieldValue.Interface())) + blankCnt
								}
							}
						}
					}
				}
			case reflect.Struct:
				t := v.Type()
				for i := 0; i < v.NumField(); i++ {
					fieldValue := v.Field(i)
					field := t.Field(i)
					maxlen[field.Name] = len(fmt.Sprintf("%s", fieldValue.Interface())) + blankCnt
				}
			}
			return maxlen

		},
		"clusterLength": func(m ...map[string]int) map[string]int {
			maxlen := make(map[string]int)
			for i := 0; i < len(m); i++ {
				if i == 0 {
					maxlen = m[i]
				} else {
					for k, v := range m[i] {
						if maxlen[k] < v {
							maxlen[k] = v
						}
					}
				}
			}
			if maxlen["ExternalIP"] == 2 {
				maxlen["ExternalIP"] = maxlen["InternalIP"]
			}
			return maxlen
		},
		"format": func(cnts ...int) map[int]int {
			lens := make(map[int]int)
			for k, v := range cnts {
				lens[k] = v + 10
			}
			return lens
		},
		"total": func(m ...int) int {
			total := 8
			for _, v := range m {
				if v == 2 {
					v = 10
				}
				total = total + v
			}
			return total
		},
	}))
	var tempText = ""
	if c.command == "update-init" {
		data.Command = "Get installed configuration"
		tempText = templates.ClusterGetKubeconfigText
	}
	if c.command == "get-kubeconfig" {
		data.Command = "Get Kubeconfig"
		tempText = templates.ClusterGetKubeconfigText
	}
	if c.command == "update" {
		data.Command = updateType
		tempText = templates.ClusterUpdateText
	}
	temp, err := koreonctlText.Parse(tempText)
	if err != nil {
		logger.Errorf("Template has errors. cause(%s)", err.Error())
		return err
	}

	// TODO: 진행상황을 어떻게 클라이언트에 보여줄 것인가?
	var buff bytes.Buffer
	err = temp.Execute(&buff, data)
	if err != nil {
		logger.Errorf("Template execution failed. cause(%s)", err.Error())
		return err
	}
	if !utils.CheckUserInput(buff.String(), "y") {
		fmt.Println("nothing to changed. exit")
		os.Exit(1)
	}

	b, err := json.Marshal(koreonToml)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}
	if err := json.Unmarshal(b, &c.extravars); err != nil {
		logger.Fatal(err.Error())
		os.Exit(1)
	}

	if c.command == "update-init" {
		currTime := time.Now()

		if err == nil {
			fmt.Println("Previous " + koreOnConfigFileName + " file exist and it will be backup")
			e := os.Rename(koreOnConfigFilePath, koreOnConfigFilePath+"_"+currTime.Format("20060102150405"))
			if e != nil {
				logger.Fatal(e)
			}
		}
	}

	ansiblePlaybookConnectionOptions := &options.AnsibleConnectionOptions{
		PrivateKey: c.privateKey,
		User:       c.user,
	}

	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Inventory: c.inventory,
		Verbose:   c.verbose,
		Tags:      c.tags,
		ExtraVars: c.extravars,
	}

	playbook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         c.playbookFiles,
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		Exec: execute.NewDefaultExecute(
			execute.WithTransformers(
				results.Prepend("Update Cluster"),
			),
		),
	}

	options.AnsibleForceColor()

	err = playbook.Run(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func checkKubeconfig(ip []string, host string) bool {
	for _, v := range ip {
		if strings.Contains(host, v) {
			return true
		}
	}

	return false
}

func getNodeIP(ip []string) map[int]string {
	nodeIPs := make(map[int]string)
	for k, v := range ip {
		nodeIPs[k] = v
	}
	return nodeIPs
}
