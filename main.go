package main

import (
	"bytes"
	"fmt"
	"github.com/manifoldco/promptui"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type option struct {
	Option string
}

var vmDir string
var dir []os.FileInfo

func main() {
	checks()
	checkVirtualMachineDirectory()
	selection := selectOption()

	if selection == "Start a virtual machine" {
		startVirtualMachine()
	} else if selection == "Stop a virtual machine" {
		stopVirtualMachine()
	} else if selection == "List all running virtual machines" {
		listRunningVMs()
	}
}

func checks() {
	// Checking if VMware is installed
	if _, err := os.Stat("/Applications/VMware Fusion.app"); os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Checking if vmrun is available
	cmd := exec.Command("vmrun")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func checkVirtualMachineDirectory() {
	if value, ok := os.LookupEnv("VIRTUAL_MACHINES_DIR"); ok {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		vmDir = homeDir + value

		if _, err := os.Stat(vmDir); os.IsNotExist(err) {
			log.Fatal(err)
		}

		dir, err = ioutil.ReadDir(vmDir)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		vmDir = homeDir + "/Virtual Machines.localized"

		if _, err := os.Stat(vmDir); os.IsNotExist(err) {
			log.Fatal(err)
		}

		dir, err = ioutil.ReadDir(vmDir)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func selectOption() string {
	options := []option{
		{Option: "Start a virtual machine"},
		{Option: "Stop a virtual machine"},
		{Option: "List all running virtual machines"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U0001F4BE{{ .Option | cyan }}",
		Inactive: "  {{ .Option | cyan }}",
		Selected: "\U0001F4BE {{ .Option | red | cyan}}",
	}

	prompt := promptui.Select{
		Label:     ">>",
		Items:     options,
		Templates: templates,
		Size:      3,
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Promt failed %v\n", err)
	}

	return options[i].Option
}

func startVirtualMachine() {
	if len(dir) < 1 {
		fmt.Println("No existing VMs")
		return
	}

	var listOfVMs []string

	for _, f := range dir {
		if filepath.Ext(f.Name()) == ".vmwarevm" {
			listOfVMs = append(listOfVMs, f.Name())
		}
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U0001F4BE{{ . | cyan }}",
		Inactive: "  {{ . | cyan }}",
		Selected: "\U0001F4BE {{ . | red | cyan}}",
	}

	prompt := promptui.Select{
		Label:     ">>",
		Items:     listOfVMs,
		Templates: templates,
		Size:      5,
	}

	_, ff, err := prompt.Run()

	if err != nil {
		fmt.Printf("Promt failed %v\n", err)
	}

	for _, f := range dir {
		if f.Name() == ff {
			dir2, err := ioutil.ReadDir(vmDir + "/" + f.Name())
			if err != nil {
				log.Fatal(err)
			}

			for _, d := range dir2 {
				if filepath.Ext(d.Name()) == ".vmx" {
					fullPath := vmDir + "/" + f.Name()
					vmxImage := d.Name()

					vmrunPath, _ := exec.LookPath("vmrun")
					cmdRun := &exec.Cmd{
						Path:   vmrunPath,
						Args:   []string{vmrunPath, "-T", "fusion", "start", vmxImage, "nogui"},
						Dir:    fullPath,
						Stdout: os.Stdout,
						Stderr: os.Stderr,
					}

					if err := cmdRun.Run(); err != nil {
						fmt.Println("Error:", err)
					}
				}
			}
		}
	}
}

func listRunningVMs() {
	vmrunPath, _ := exec.LookPath("vmrun")
	cmdRun := &exec.Cmd{
		Path:   vmrunPath,
		Args:   []string{vmrunPath, "list"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := cmdRun.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}

func runningVMs() []string {
	vmrunCmd := exec.Command("vmrun", "list")
	awkCmd := exec.Command("awk", "/.vmx/,0")

	awkCmd.Stdin, _ = vmrunCmd.StdoutPipe()

	var buf bytes.Buffer
	var buff strings.Builder

	awkCmd.Stdout = &buf

	awkCmd.Start()
	vmrunCmd.Run()
	awkCmd.Wait()

	io.Copy(&buff, &buf)

	return strings.Split(buff.String(), "\n")
}

func truncatedRunningVMs() []string {
	vmrunCmd := exec.Command("vmrun", "list")
	awkCmd := exec.Command("awk", "/.vmx/,0")
	sedCmd := exec.Command("sed", "s:.*/::")

	awkCmd.Stdin, _ = vmrunCmd.StdoutPipe()
	sedCmd.Stdin, _ = awkCmd.StdoutPipe()

	var buf bytes.Buffer
	var buff strings.Builder

	sedCmd.Stdout = &buf

	sedCmd.Start()
	awkCmd.Start()
	vmrunCmd.Run()
	awkCmd.Wait()
	sedCmd.Wait()

	io.Copy(&buff, &buf)

	return strings.Split(buff.String(), "\n")
}

func stopVirtualMachine() {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U0001F4BE{{ . | cyan }}",
		Inactive: "  {{ . | cyan }}",
		Selected: "\U0001F4BE {{ . | red | cyan}}",
	}

	prompt := promptui.Select{
		Label:     ">>",
		Items:     truncatedRunningVMs(),
		Templates: templates,
		Size:      5,
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Promt failed %v\n", err)
	}

	for ii, vm := range runningVMs() {
		if i == ii {
			vmrunPath, _ := exec.LookPath("vmrun")
			cmdRun := &exec.Cmd{
				Path:   vmrunPath,
				Args:   []string{vmrunPath, "-T", "fusion", "stop", vm},
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}

			if err := cmdRun.Run(); err != nil {
				fmt.Println("Error:", err)
			}
		}
	}
}
