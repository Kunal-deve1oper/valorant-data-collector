# Valorant Data Collector

A Go-based application for collecting and processing Valorant game data.

## Prerequisites

- [Go](https://golang.org/dl/) (version 1.20 or later recommended)
- [Git](https://git-scm.com/)

## Installation and Setup

1. **Clone the repository:**

   ```bash
   git clone [https://github.com/Kunal-deve1oper/valorant-data-collector.git](https://github.com/Kunal-deve1oper/valorant-data-collector.git)
   cd valorant-data-collector
   ```

2. **Initialize dependencies:**

   ```bash
   go mod download
   ```

3. **Configure environment variables:**

   create a `.env` file in the root directory:

   ```bash
   add API_KEY=<your api key>
   ```

4. **Get api key
    get api key from https://discord.com/channels/704231681309278228/1451136198796906587/1476477184939131064

## Running the Project

To run the application directly from the source:

```bash
go run .
```

To build the executable binary:

```bash
go build -o collector main.go
./collector
```
