package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/crypto/ssh"
	yaml "gopkg.in/yaml.v3"
)

// server key
var hostKey ssh.PublicKey

func work(host string, action string) {
	// An SSH client is represented with a ClientConn.

	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile(".ssh/key")
	if err != nil {
		log.Fatalf("Remote action: unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Remote action: unable to parse private key: %v", err)
	}

	// To authenticate with the remote server you must pass at least one
	// implementation of AuthMethod via the Auth field in ClientConfig,
	// and provide a HostKeyCallback.
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		//HostKeyCallback: ssh.FixedHostKey(hostKey),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Fatal("Remote action: Failed to dial: ", err)
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Remote action: Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(action); err != nil {
		log.Println("Remote action:  failed to run " + err.Error())
	} else {
		log.Println("Remote action: ", b.String())
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello from a AutoHeal\n")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var grafanaAlert map[string]interface{}
	json.Unmarshal(body, &grafanaAlert)

	msg := grafanaAlert["message"].(string)
	rule := grafanaAlert["ruleName"].(string)
	state := grafanaAlert["state"].(string)

	log.Println("Message:", msg, " Rule:", rule, " State: ", state)
	go work(msg, rule)
}

type healer struct {
	Server healerSrv `yaml:"server"`
}

type healerSrv struct {
	Addr string `yaml:"addr"`
	Port string `yaml:"port"`
}

type service struct {
	SSHEntries []sshEntry `yaml:"service"`
}

type sshEntry struct {
	Name    string `yaml:"name"`
	Command sshCmd `yaml:"command"`
}

type sshCmd struct {
	Addr string `yaml:"addr"`
	Port int    `yaml:"port"`
	Cmd  string `yaml:"exe"`
}

func main() {
	var config healer
	var cfg service

	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatal("Error open config file ", err)
	}

	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	// do some work with cfg

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	addr := config.Server.Addr
	port := config.Server.Port

	http.HandleFunc("/", handler)

	if err := http.ListenAndServe(addr+":"+port, nil); err != nil {
		log.Fatal("Server start failed..." + err.Error())
	}

}
