# MarketFlow

MarketFlow is a real-time market data processing system built in Go using hexagonal architecture. The application collects data from cryptocurrency exchange simulators or generates test data, aggregates prices, stores them in PostgreSQL, and caches them in Redis. A built-in REST API provides convenient access to aggregated information.

## 🚀 Live Demo

View Live Application
(replace with actual link if available)

## 🛠️ Technologies Used

Backend: Go (1.21+)

Database: PostgreSQL

Cache: Redis

Deployment: Docker, Docker Compose

## ✨ Features

Real-time aggregation of market prices

Live/Test modes for flexible data sources

Worker pool for concurrent feed processing (5 workers per exchange)

Fan-In / Fan-Out architecture for data streams

Batch insertion to PostgreSQL for efficiency

Automatic fallback to DB if Redis is unavailable

REST API to fetch latest, highest, lowest, and average prices

System health endpoint and logging

## 📦 Installation

Clone the repository:

git clone git@github.com:AikaBe/marketflow-System.git
cd marketflow


Load exchange images (if using simulators):

docker-compose run --rm load_images


Build the project:

docker-compose build


Configure config.yaml:

postgres:
host: localhost
port: 5432
user: marketflow
password: secret
dbname: marketflow_db

redis:
host: localhost
port: 6379
password: ""

exchanges:
- name: exchange1
  host: 127.0.0.1
  port: 40101
- name: exchange2
  host: 127.0.0.1
  port: 40102
- name: exchange3
  host: 127.0.0.1
  port: 40103

## 🎯 Usage
Run the application with Docker Compose:
docker-compose up

Running exchange simulators (Live Mode):
docker load -i exchange1_amd64.tar
docker run -p 40101:40101 -d exchange1_amd64

docker load -i exchange2_amd64.tar
docker run -p 40102:40102 -d exchange2_amd64

docker load -i exchange3_amd64.tar
docker run -p 40103:40103 -d exchange3_amd64

API Examples

Fetch latest price for BTCUSDT:

curl http://localhost:8080/prices/latest/BTCUSDT


Switch to test mode:

curl -X POST http://localhost:8080/mode/test


Check system health:

curl http://localhost:8080/health

## 🏗️ Architecture

MarketFlow is built using Hexagonal Architecture (Ports & Adapters):

Domain Layer: business logic and models

Application Layer: use-case processing and data flow management

Adapters:

HTTP Adapter (REST API)

Storage Adapter (PostgreSQL)

Cache Adapter (Redis)

Exchange Adapter (Live/Test sources)

Features include worker pool processing, fan-in/fan-out architecture, batch inserts, Redis caching with fallback, and failover for resilient data fetching.

## 🔮 Future Improvements

Add WebSocket endpoint for live price streaming

Support additional trading pairs and exchanges

Add authentication and user management for API access

Implement historical data analytics and visualization