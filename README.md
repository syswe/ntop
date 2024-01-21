# 🚀 ntop - Kubernetes Node & Pod Metrics Tool
![ntop-logo](https://i.ibb.co/WzWRNcz/ntop-logo.png)

**ntop** is an open-source tool designed to provide clear and visual insights into the status of nodes and pods in your Kubernetes cluster. Leveraging Kubernetes' metrics server, ntop fetches and displays real-time CPU and memory utilization metrics in an easy-to-consume command-line interface.

![GitHub release (with filter)](https://img.shields.io/github/v/release/devopswe/ntop?style=for-the-badge) ![GitHub contributors](https://img.shields.io/github/contributors/devopswe/ntop?style=for-the-badge) ![GitHub Repo stars](https://img.shields.io/github/stars/devopswe/ntop?style=for-the-badge)

---

## 📚 Table of Contents
1. [Installation](#installation)
2. [Usage](#usage)
3. [Features](#features)
4. [What's New in 0.2.0](#whats-new)
5. [Development](#development)
6. [Roadmap](#roadmap)
7. [Contributing](#contributing)
8. [License](#license)

## 🛠 Installation

### Prerequisites:
- `Go` (1.16 or later)
- Access to a Kubernetes cluster
- Kubernetes configuration file (usually at `~/.kube/config`)

### Build:
Clone the repository and build the tool:

```bash
git clone https://github.com/devopswe/ntop
cd ntop
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ntop-linux-64-0.2.0
```

## 🖥 Usage

Ensure the metrics-server is installed on your cluster:

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

> **Note**: If you encounter SSL/TLS errors, modify `components.yaml` to include the `--kubelet-insecure-tls` flag.

Run ntop:

```bash
./ntop-linux-64-0.2.0 --kubeconfig=path/to/your/kubeconfig
```

Optional flags:

- `--pods`: Switch to pods view.
- `--countpods=<number>`: Specify the number of pods to display (default is 10).

### Listing Nodes Example:
![Example nodes listing of NTOP 0.2.0](nodes-list.png)

### Listing Top Pods Example:
![Example pods listing of NTOP 0.2.0](top-pods-list.png)

## ⭐ Features

- Real-time CPU and memory usage metrics for nodes and pods.
- Toggle between node and pod views with ease.
- Customize the number of pods displayed.
- Simple and intuitive command-line interface.
- Utilizes `tablewriter` for clear, tabulated displays.

## 🌟 What's New in 0.2.0

- Live table display of nodes and pods.
- `--pods` flag to easily switch to pod metrics.
- `--countpods` flag to specify the number of pods displayed.
- Improved performance and reduced waiting times.
- Instantaneous pod data retrieval with the new concurrent structure.
- Cross-platform availability: Windows, Linux, MacOS (Intel & ARM).

Special thanks to [@dbtek](https://github.com/dbtek) and [@faruktoptas](https://github.com/faruktoptas) for their contributions!

## 🛠 Development

### Dependencies:
- `k8s.io/client-go` and `k8s.io/metrics` for Kubernetes interaction.
- `github.com/olekukonko/tablewriter` for rendering terminal tables.

Use `go mod tidy` to manage dependencies.

## 🗺 Roadmap

Our vision for ntop includes:

1. **Modularization**: Refactor for maintainability and scalability.
2. **`kubectl` Plugin**: Enhance accessibility and integration.
3. **Enhanced Filtering**: Customize data display based on user preferences.
4. **Trend Analysis**: Introduce historical data tracking and analysis.
5. **Interactive UI**: Develop a terminal-based interactive UI.
6. **Support Additional Resources**: Extend monitoring to more Kubernetes resources.

## 🤝 Contributing

Contributions are welcome! Please follow the standard PR process for your contributions.

## 📜 License

Licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.

---

Enhance your Kubernetes monitoring with **ntop**. Try it out and let us know your thoughts!