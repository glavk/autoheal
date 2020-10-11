package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/crypto/ssh"
)

// server key
var hostKey ssh.PublicKey

func work(host string, action string) {
	// An SSH client is represented with a ClientConn.

	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile("/home/sea/.ssh/rhce")
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
	defer r.Body.Close() // важный пункт!
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var grafanaAlert map[string]interface{}
	json.Unmarshal(body, &grafanaAlert)
	//fmt.Printf("unpacked in empty interface:\n%#v\n\n", grafanaAlert)
	msg := grafanaAlert["message"].(string)
	rule := grafanaAlert["ruleName"].(string)
	state := grafanaAlert["state"].(string)

	log.Println("Message:", msg, " Rule:", rule, " State: ", state)
	go work(msg, rule)
}

func main() {
	http.HandleFunc("/", handler)

	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Fatal("Server start failed..." + err.Error())
	}
}
