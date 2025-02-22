# NTLM Hunter

NTLM Hunter is a tool designed to scan a list of hosts for NTLM authentication endpoints and collect NTLM challenge responses. It's particularly useful for auditing and testing NTLM authentication deployments and ensuring secure configurations.

---

## Features

- **Port Scanning**: Tests predefined ports (e.g., 25, 80, 443, 445, 8080, 8443, 3389) for availability.
- **Endpoint Discovery**: Scans multiple NTLM-related paths on supported services such as HTTP, HTTPS, SMB, SMTP, and RDP.
- **NTLM Challenge Extraction**: Attempts to retrieve NTLM challenges from endpoints and displays the response.
- **Concurrent Execution**: Uses Goroutines and WaitGroups for parallel scanning and lookups to optimize performance.

---

## Prerequisites

1. **Go**: Installed on your system (version 1.22 or later recommended).
2. **External Dependency**: The program requires the `github.com/bogey3/NTLM_Info` library. Make sure it's installed and available in your Go environment.

---

## Installation

1. Clone the repository or copy the necessary files.
2. Install dependencies:

    ```bash
    go get github.com/bogey3/NTLM_Info
    ```

3. Build the program using the following command:

    ```bash
    go build -o ntlm-hunter main.go
    ```

---

## Run Instructions

1. Prepare a text file containing a list of hosts (one per line).
   Example: `hosts.txt`
   ```
   example.com
   testserver.local
   ```

2. Run the executable with the host file as an argument:

    ```bash
    ./ntlm-hunter hosts.txt
    ```

---

## How it Works

1. **Input**: The program reads a list of hostnames or IP addresses from a file provided as a command-line argument.
2. **Port Testing**: It scans a set of predefined ports to check if they are open.
3. **Generate URLs**: For each open port, the program generates URLs using common NTLM-related paths.
4. **NTLM Challenge Lookup**: For each generated URL, the program attempts to retrieve an NTLM challenge using the `NTLM_Info` library.
5. **Output**: The program prints information about endpoints that responded successfully with NTLM challenges, along with challenge details.

---

## Configuration

The program includes predefined ports and NTLM paths:

- **Ports**:
   - `25` (SMTP)
   - `80` (HTTP)
   - `443` (HTTPS)
   - `445` (SMB)
   - `8080` (Alternative HTTP)
   - `8443` (Alternative HTTPS)
   - `3389` (RDP)

- **Paths**:
  Includes common NTLM authentication paths, such as:
   - `/owa/`
   - `/ews/`
   - `/rpc/`
   - `/api/`
   - `/adfs/ls/`
   - And several more

You can customize the list of ports or paths by editing the `ports` and `ntlmPaths` variables in the code directly.

---