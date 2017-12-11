package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"golang.org/x/sys/unix"
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

	url := fmt.Sprintf("https://%s:%s/puppet/v3/status/test?environment=%s", server, port, environment)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	request.Header.Set("Accept", "application/json")

	response, err := httpClient().Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode == 404 {
		return false
	} else if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Printf("Unexpected status %q checking environment %q. Keeping environment.", response.Status, environment)
	}

	return true
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

	// unix.Flock is not supported on Solaris. This call does not block.
	lock := unix.Flock_t{
		Type:   unix.F_WRLCK,
		Whence: 0, // There's no constant for this
		Start:  0,
		Len:    0, // to end of file
	}
	err = unix.FcntlFlock(uintptr(file.Fd()), unix.F_SETLK, &lock)
	if err == unix.EAGAIN {
		// Another process has the lock
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
