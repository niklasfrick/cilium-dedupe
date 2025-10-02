# Cilium Flow Deduplication Tool

A Go utility that processes Cilium flow logs to extract unique network flows based on source identity, destination identity, and destination port. This tool helps simplify the creation of network policies by providing a deduplicated JSON output that can be easily imported into the Cilium Network Policy Editor.

## Overview

This tool reads Cilium flow log files (`.log` files) from the current directory, parses the JSON flow data, and creates a deduplicated output file containing only unique flows. The deduplication is based on:
- Source identity
- Destination identity  
- Destination port (TCP)

## Prerequisites

- Go 1.21 or later
- Cilium cluster with flow logging enabled

## Installation

1. Clone this repository:
```bash
git clone <repository-url>
cd cilium-dedupe
```

2. Build the tool:
```bash
go build -o cilium-dedupe get-unique-flows.go
```

## Usage

1. Place your Cilium flow log files (`.log` files) in the same directory as the tool
2. Run the tool:
```bash
./cilium-dedupe
```

The tool will:
- Process all `.log` files in the current directory
- Extract unique flows based on source/destination identity and port
- Generate a timestamped output file: `unique-flows-YYYYMMDD-HHMMSS.json`
- Display progress information to stderr

### Example Output

```
Processing file: cilium-flows.log
Total unique flows: 1,247
Unique flows written to: unique-flows-20241201-143022.json
```

## Configuring Cilium for Flow Logging

To enable Cilium to automatically export flows to `.log` files, you need to configure flow logging in your Cilium installation.

### Method 1: Using Cilium CLI (Recommended)

1. Enable Hubble with flow logging:
```bash
cilium hubble enable --flow-logging
```

2. Configure Hubble to write logs to files:
```bash
cilium hubble config set --flow-logging-file-path /var/log/cilium/flows.log
```

### Method 2: Using Helm Values

Add the following to your `values.yaml`:

```yaml
hubble:
  enabled: true
  flowLogging:
    enabled: true
    filePath: "/var/log/cilium/flows.log"
    rotation:
      enabled: true
      maxSize: 100  # MB
      maxAge: 7     # days
      maxBackups: 5
```

Then upgrade your Cilium installation:
```bash
helm upgrade cilium cilium/cilium -f values.yaml
```

### Method 3: Manual Configuration

1. Edit the Cilium ConfigMap:
```bash
kubectl edit configmap cilium-config -n kube-system
```

2. Add the following configuration:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cilium-config
  namespace: kube-system
data:
  hubble-flow-logging-enabled: "true"
  hubble-flow-logging-file-path: "/var/log/cilium/flows.log"
```

3. Restart Cilium pods:
```bash
kubectl rollout restart daemonset/cilium -n kube-system
```

## Log File Collection

After enabling flow logging, you can collect the log files from your Cilium pods:

```bash
# Get the log files from all Cilium pods
kubectl cp kube-system/cilium-<pod-name>:/var/log/cilium/flows.log ./cilium-flows.log

# Or collect from multiple pods
for pod in $(kubectl get pods -n kube-system -l k8s-app=cilium -o jsonpath='{.items[*].metadata.name}'); do
  kubectl cp kube-system/$pod:/var/log/cilium/flows.log ./flows-$pod.log
done
```

## Using with Cilium Network Policy Editor

1. Run the deduplication tool on your collected log files
2. Open the generated `unique-flows-*.json` file
3. Copy the relevant flow entries
4. Import them into the [Cilium Network Policy Editor](https://editor.cilium.io/) or use them as reference for creating network policies

### Example Flow Entry

The tool processes entries like this:
```json
{
  "flow": {
    "time": "2025-02-26T15:51:21.102907649Z",
    "verdict": "AUDIT",
    "ethernet": {
      "source": "aa:bb:cc:11:22:33",
      "destination": "dd:ee:ff:44:55:66"
    },
    "IP": {
      "source": "10.244.8.160",
      "destination": "10.244.10.71",
      "ipVersion": "IPv4"
    },
    "l4": {
      "TCP": {
        "source_port": 51004,
        "destination_port": 5432,
        "flags": { "SYN": true }
      }
    },
    "source": {
      "namespace": "harbor",
      "pod_name": "harbor-core-7cfb96b999-cnw9r",
      ...
    },
    "destination": {
      "namespace": "harbor",
      "pod_name": "harbor-database-0",
      ...
    },
    "traffic_direction": "INGRESS",
    "Summary": "TCP Flags: SYN"
    ...
  }
  ...
}

```

## Troubleshooting

### No log files found
- Ensure you have `.log` files in the current directory
- Check that Cilium flow logging is properly configured
- Verify log file permissions

### Parsing errors
- The tool skips malformed JSON entries automatically
- Check that your log files contain valid JSON flow data
- Ensure the log format matches the expected structure

### Large log files
- The tool uses a 512KB buffer for reading large lines
- For very large files, consider splitting them or processing in batches

## License

This project is licensed under the MIT License - see the LICENSE file for details.