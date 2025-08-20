# AI-Powered Code Review Assistant

[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Code Coverage](https://img.shields.io/badge/coverage-70%25-brightgreen)](https://github.com/)

An intelligent, automated code review system that leverages Large Language Models (LLMs) to analyze pull requests, provide context-aware feedback, and improve code quality seamlessly within your Git workflow.

## Overview

The traditional code review process, while crucial, can be a significant bottleneck in development cycles. This project automates a large part of this process by using an AI assistant to perform initial reviews on pull requests.

When a PR is opened in a repository (e.g., on GitHub), the system is notified via a webhook. It then clones the repository, parses the entire codebase to understand its structure, and creates vector embeddings of the code. Using a **Retrieval-Augmented Generation (RAG)** approach, it finds the most relevant code snippets to provide context for the changes in the PR. This context, along with the code diff, is sent to an LLM, which generates a review and posts it as a comment directly on the PR.

## Key Features

-   **🤖 Automated PR Analysis**: Triggers automatically on pull request events from version control systems like GitHub or Gitlab.
-   **💡 Context-Aware Reviews**: Goes beyond simple line-by-line analysis. By vectorizing the entire codebase, it understands the broader context of any change, leading to more insightful suggestions.
-   **🧠 LLM-Powered Intelligence**: Utilizes the power of Large Language Models to check for potential bugs, suggest style improvements, and ensure best practices are followed.
-   **🔗 Seamless Integration**: Posts comments directly on the pull request, fitting naturally into the existing developer workflow.
-   **🚀 Scalable & Resilient by Design**: Built on a microservices architecture with a Kafka message queue, ensuring high availability and the ability to process numerous PRs concurrently.
-   **🔧 Extensible**: Designed with interfaces and a modular structure, making it easy to add support for new programming languages or version control systems.

## System Architecture

The system is designed as a distributed, event-driven application composed of two primary microservices that communicate asynchronously via a Kafka message broker. This decouples the initial event ingestion from the intensive processing logic, enhancing scalability and reliability.



1.  **API Gateway**: This service acts as the public-facing ingress point. It's responsible for receiving and validating webhooks from version control systems (e.g., GitHub). Upon successful validation, it converts the payload into a standardized internal event format and publishes it to a Kafka topic. It only acknowledges the webhook after the event is successfully queued, guaranteeing no requests are lost.

2.  **Code Reviewer Service**: This is the core engine of the system. A pool of workers consumes events from the Kafka topic. For each event, it performs the full code review pipeline:
    1.  **Clone** the repository and download the PR diff.
    2.  **Parse** the entire codebase using Tree-sitter for accurate, syntax-aware chunking of code into functions, classes, etc.
    3.  **Embed & Index** these chunks into a ChromaDB vector store.
    4.  **Retrieve & Generate**: Embed the PR diff, find the most relevant code chunks from ChromaDB as context, and send everything to the LLM to generate the review.
    5.  **Comment**: Post the LLM's response back to the original pull request.

## Technology Stack

| Category                  | Technology                           |
| ------------------------- |--------------------------------------|
| **Backend** | Golang                               |
| **API Framework** | Gin                                  |
| **Messaging Broker** | Apache Kafka                         |
| **Vector Database** | ChromaDB                             |
| **AI / LLM Orchestration**| LangChainGo, OpenAI                  |
| **Containerization** | Docker, Docker Compose, Docker Swarm |
| **Observability** | Prometheus, Grafana                  |

## 🚀 Getting Started

You can run the entire system locally using Docker Compose.

### Prerequisites

-   Docker and Docker Compose
-   Go (for running tests)

### Local Development

1.  **Clone the repository:**
    ```sh
    git clone <your-repo-url>
    cd <your-repo-directory>
    ```

2.  **Configure Environment Variables:**
    Each service (`api-gateway` and `code-reviewer`) has its own `.env` file. Copy the example files and fill in your secrets:
    ```sh
    cp services/api-gateway/.env.example services/api-gateway/.env
    cp services/code-reviewer/.env.example services/code-reviewer/.env
    ```
    You will need to provide your `GITHUB_ACCESS_TOKEN`, `LLM_OPEN_AI_API_KEY`, and a `GITHUB_WEBHOOK_SECRET`.

3.  **Run the System:**
    ```sh
    docker-compose up --build
    ```
    This will start all services, including Kafka, ChromaDB, Prometheus, and Grafana. The API Gateway will be available at `http://localhost:8080`.

4.  **Set up a Webhook:**
    To receive events from GitHub, you'll need to expose your local API Gateway to the internet. Tools like **ngrok** are great for this.
    ```sh
    ngrok http 8080
    ```
    Use the public URL provided by ngrok (e.g., `https://<unique-id>.ngrok.io`) to set up a webhook in your GitHub repository's settings. The endpoint is `/github-webhook`.

---

## Testing

The project has a comprehensive test suite covering unit, integration, and end-to-end scenarios. To run all tests:
```sh
go test ./...