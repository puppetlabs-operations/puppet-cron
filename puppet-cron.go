package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const Puppet = "/opt/puppetlabs/bin/puppet"
const LockPath = "/var/run/puppet-cron.lock"

func puppetConfigSectionGet(section string, key string) string {
	args := []string{"config", "print"}

	if section != "" {
		args = append(args, "--section", section)
	}

	args = append(args, key)
	output, err := exec.Command(Puppet, args...).Output()
	if err != nil {
		log.Fatalf("Failed: %s %s", Puppet, strings.Join(args, " "))
	}

	return string(output[:len(output)-1])
}

func puppetConfigGet(key string) string {
	return puppetConfigSectionGet("", key)
}

func puppetConfigSectionSet(section string, key string, value string) {
	args := []string{"config", "set", "--section", section, key, value}

	_, err := exec.Command(Puppet, args...).Output()
	if err != nil {
		log.Fatalf("Failed: %s %s", Puppet, strings.Join(args, " "))
	}
}

func httpClient() *http.Client {
	certname := puppetConfigGet("certname")
	ssldir := puppetConfigGet("ssldir")

	clientCertPath := fmt.Sprintf("%s/certs/%s.pem", ssldir, certname)
	clientKeyPath := fmt.Sprintf("%s/private_keys/%s.pem", ssldir, certname)
	caCertPath := fmt.Sprintf("%s/certs/ca.pem", ssldir)

	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}

func isValidEnvironment(environment string) bool {
	server := puppetConfigSectionGet("agent", "server")
	port := puppetConfigSectionGet("agent", "masterport")
	client := httpClient()

	url := fmt.Sprintf("https://%s:%s/puppet/v3/environments", server, port)
	response, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var object map[string]interface{}
	err = json.Unmarshal(body, &object)
	if err != nil {
		log.Fatal(err)
	}

	validEnvironments := object["environments"].(map[string]interface{})
	_, ok := validEnvironments[environment]

	return ok
}

func obtainLock() {
	err := os.MkdirAll(filepath.Dir(LockPath), os.FileMode(0755))
	if err != nil {
		log.Panic(err)
	}

	// This has to be 0600 so that a non-priviledged user can't deny access.
	file, err := os.OpenFile(LockPath, os.O_RDWR|os.O_CREATE, os.FileMode(0600))
	if err != nil {
		log.Panic(err)
	}

	// Do not close the file; we want to keep the lock active until puppet has
	// finished running.

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == syscall.EWOULDBLOCK {
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			log.Fatal("Another process is running. Could not read PID from lock file: " + err.Error())
		}
		log.Fatal("puppet-cron.py is already running with pid " + scanner.Text())
	} else if err != nil {
		log.Panic(err)
	}

	_, err = file.Seek(io.SeekStart, 0)
	if err != nil {
		log.Panic(err)
	}

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(strconv.Itoa(os.Getpid()) + `

puppet-cron does not delete this file on completion, so the PID above may no
longer represent a puppet-cron process. puppet-cron uses flock, so the lock is
automatically released on process exit.
`)
	if err != nil {
		log.Panic(err)
	}

	err = writer.Flush()
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	obtainLock()

	environment := puppetConfigSectionGet("agent", "environment")

	if environment == "" || !isValidEnvironment(environment) {
		log.Printf("Environment %q is invalid; resetting", environment)
		puppetConfigSectionSet("agent", "environment", "production")
	}

	err := syscall.Exec(Puppet, []string{"puppet", "agent", "--no-daemonize", "--onetime"}, os.Environ())
	if err != nil {
		log.Fatal(err)
	} else {
		// WTF. exec should only return if it fails
		log.Panic("syscall.Exec returned with no error")
	}
}
