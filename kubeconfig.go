package main

import (
	"fmt"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
	"sort"
)

type KubeConfig struct {
	APIVersion     string      `yaml:"apiVersion"`
	Clusters       []Clusters  `yaml:"clusters"`
	Contexts       []Contexts  `yaml:"contexts"`
	CurrentContext string      `yaml:"current-context"`
	Kind           string      `yaml:"kind"`
	Preferences    Preferences `yaml:"preferences"`
	Users          []Users     `yaml:"users"`
}
type Cluster struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}
type Clusters struct {
	Cluster Cluster `yaml:"cluster"`
	Name    string  `yaml:"name"`
}
type Context struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}
type Contexts struct {
	Context Context `yaml:"context"`
	Name    string  `yaml:"name"`
}
type Preferences struct {
}
type User struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}
type Users struct {
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

// ListContexts listet alle Kontexte der globalen kubeconfig unter /etc/k8s/config auf
func (c *KubeConfig) ListContexts() {
	activeSign := " "
	sortedContexts := c.Contexts
	sort.Slice(sortedContexts, func(i, j int) bool {
		return sortedContexts[i].Name < sortedContexts[j].Name
	})

	var oldConfig KubeConfig

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Kann Home Verzeichnis nicht ermitteln: %v", err)
	}

	localConfigPath := path.Join(homedir, localConfigPathExt)

	var currentContext string
	var currentNamespace string

	// lokale config einlesen, falls diese existiert
	if _, err = os.Stat(localConfigPath); err == nil {
		oldConfig.Load(localConfigPath)
		currentContext = oldConfig.CurrentContext
		for _, cntCon := range oldConfig.Contexts {
			if cntCon.Name == currentContext {
				currentNamespace = cntCon.Context.Namespace
				break
			}
		}
	}

	var wide bool
	if term.IsTerminal(0) {
		w, _, _ := term.GetSize(0)
		if w > 175 {
			wide = true
		}
	}

	if wide {
		fmt.Printf("%-3s %-35s %-50s %-60s %-30s\n", "CUR", "NAME", "CLUSTER", "AUTHINFO", "NAMESPACE")
	} else {
		fmt.Printf("%-3s %-35s %-50s\n", "CUR", "NAME", "CLUSTER")
	}

	var ns string
	for _, con := range sortedContexts {
		{
			if con.Name == currentContext {
				activeSign = " * "
				ns = currentNamespace
			} else {
				activeSign = "   "
				ns = ""
			}
			if wide {
				fmt.Printf("%-3s %-35s %-50s %-60s %-30s\n", activeSign, con.Name, con.Context.Cluster, con.Context.User, ns)
			} else {
				fmt.Printf("%-3s %-35s %-50s\n", activeSign, con.Name, con.Context.Cluster)
			}
		}
	}
}

func (c *KubeConfig) Load(configpath string) {
	srcFile, err := os.ReadFile(configpath)
	if err != nil {
		log.Printf("Kann kubeconfig %s nicht lesen: %v", configpath, err)
	}

	err = yaml.Unmarshal(srcFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}

func (c *KubeConfig) Save(configpath string) {
	out, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	dir := path.Dir(configpath)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		log.Fatalf("Fehler beim erstellen des %s Verzeichnisses: %v", dir, err)
	}

	err = os.WriteFile(configpath, out, 0600)
	if err != nil {
		log.Fatalf("Kann kubeconfig %s nicht schreiben: %v", configpath, err)
	}
}
