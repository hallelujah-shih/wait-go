package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var appVersionFile = "./VERSION"

func readVersion(appVersionFile string) string {
	dat, _ := ioutil.ReadFile(appVersionFile)
	return string(dat)
}

// PathDetector detects if binaries are in path
type PathDetector interface {
	inPath(command string) bool
}
type localPathDetector struct {
}

func (pathDetector localPathDetector) inPath(command string) bool {
	_, err := exec.LookPath(command)
	if err != nil {
		return false
	}
	return true
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "String"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func waitFor(waitsFlags arrayFlags, commandFlags arrayFlags, timeoutFlag int, intervalFlag int, shell string) {
	for _, wait := range waitsFlags {
		processWait(wait, timeoutFlag, intervalFlag, shell)
	}

	for _, command := range commandFlags {
		processCommandExec(command, timeoutFlag, intervalFlag, shell)
	}

}

func checkTcpDial(host, port string, intervalFlag int) {
	for {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Duration(intervalFlag)*time.Second)
		if err != nil {
			log.Println(err)
			log.Printf("Sleeping %d seconds waiting for host\n", intervalFlag)
			time.Sleep(time.Duration(intervalFlag) * time.Second)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}
}

func checkShell(shellCmd, shell string, intervalFlag int) {
	for {
		out, err := exec.Command(shell, "-c", shellCmd).Output()
		if err != nil {
			log.Printf("Sleeping %d seconds waiting for command - %s - to return\n", intervalFlag, shellCmd)
			time.Sleep(time.Duration(intervalFlag) * time.Second)
		} else {
			log.Println(string(out))
			break
		}
	}
}

func checkHTTP(u *url.URL, intervalFlag int) {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 30 * time.Second,
	}

	for {
		rsp, err := client.Get(u.String())
		if err != nil {
			log.Println(err)
			log.Printf("Sleeping %d seconds waiting for http\n", intervalFlag)
			time.Sleep(time.Duration(intervalFlag) * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(rsp.Body)
		rsp.Body.Close()

		if rsp.StatusCode == 200 {
			break
		}

		log.Println("status:", rsp.Status, "body:", string(body))
		log.Printf("Sleeping %d seconds waiting for http\n", intervalFlag)
		time.Sleep(time.Duration(intervalFlag) * time.Second)
	}
}

func checkIsUrl(urlPath string) (bool, *url.URL) {
	u, err := url.Parse(urlPath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return true, u
	}
	return false, nil
}

func processWait(wait string, timeoutFlag int, intervalFlag int, shell string) {
	isUrl, u := checkIsUrl(wait)
	if isUrl {
		switch strings.ToLower(u.Scheme) {
		case "http":
			fallthrough
		case "https":
			checkHTTP(u, intervalFlag)
		default:
			log.Panicln("not support scheme:", u.Scheme)
		}
		return
	}

	pattern, _ := regexp.Compile("(.*):(.*)")
	matches := pattern.FindAllStringSubmatch(wait, -1)
	if len(matches) > 0 {
		host := matches[0][1]
		port := matches[0][2]
		checkTcpDial(host, port, intervalFlag)
		return
	}

	checkShell(wait, shell, intervalFlag)
}

func processCommandExec(command string, timeoutFlag int, intervalFlag int, shell string) {
	cmd := exec.Command(shell, "-c", command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("cmd:", command, "stdout_pipe err:", err)
	} else {
		go io.Copy(os.Stdout, stdout)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalln("cmd:", command, "stderr_pipe err:", err)
	} else {
		go io.Copy(os.Stderr, stderr)
	}

	err = cmd.Start()
	if err != nil {
		log.Printf("Sleeping %d seconds waiting for command - %s - to return\n", intervalFlag, command)
		time.Sleep(time.Duration(intervalFlag) * time.Second)
	}
	cmd.Wait()
}

func chooseShell(pathDetector PathDetector) string {
	if pathDetector.inPath("bash") {
		return "bash"
	} else if pathDetector.inPath("sh") {
		return "sh"
	} else {
		panic("Neither bash or sh present on system")
	}
}

func mainExecution(waitsFlags arrayFlags, commandFlags arrayFlags, timeoutFlag int, intervalFlag int, version bool, pathDetector localPathDetector) int {
	if version {
		log.Println(readVersion(appVersionFile))
		return 0
	}

	if len(waitsFlags) == 0 || len(commandFlags) == 0 {
		log.Println("You must specify at least a wait and a command. Please see --help for more information.")
		return 1
	}
	shell := chooseShell(pathDetector)
	waitFor(waitsFlags, commandFlags, timeoutFlag, intervalFlag, shell)
	return 2
}

func main() {
	var pathDetector = localPathDetector{}
	// Set custom logger
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	var waitsFlags arrayFlags
	var commandFlags arrayFlags

	flag.Var(&waitsFlags, "wait", "You can specify the HOST and TCP PORT using the format HOST:PORT, or http[s]://domain/path?args, or you can specify a command that should return an output. Multiple wait flags can be added.")
	flag.Var(&commandFlags, "command", "Command that should be run when all waits are accessible. Multiple commands can be added.")
	timeoutFlag := flag.Int("timeout", 600, "Timeout untill script is killed.")
	intervalFlag := flag.Int("interval", 15, "Interval between calls")
	version := flag.Bool("version", false, "Prints current version")
	flag.Parse()
	returnValue := mainExecution(waitsFlags, commandFlags, *timeoutFlag, *intervalFlag, *version, pathDetector)
	os.Exit(returnValue)
}
