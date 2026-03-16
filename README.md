# Full-Stack Go Todo Application 🚀

A robust, containerized Task Management System built with a focus on
**Clean Architecture**, security, and DevOps best practices.

## 🏗 Architecture & Tech Stack

This project is a complete full-stack solution featuring:

-   **Backend**: Developed in **Go (Golang)** using a modular structure
    (`handlers`, `middleware`, `models`).
-   **Database**: **PostgreSQL 16** for reliable persistent storage.
-   **Frontend**: A responsive Single Page Application (SPA) built with
    **Vanilla JavaScript** and **Bootstrap 5**.
-   **Security**:
    -   **JWT (JSON Web Tokens)** for secure authentication
    -   **bcrypt** for password hashing
-   **Infrastructure**:
    -   **Docker & Docker Compose** with multi-stage builds for
        ultra-lightweight images (\~20MB)
    -   **Nginx** as a reverse proxy and static file server
    -   **GitHub Actions** for automated CI/CD deployment to a
        self-hosted Proxmox environment

## 📁 Project Structure

``` text
├── cmd/server/           # Main entry point (initialization & routing)
├── internal/             # Private application logic
│   ├── handlers/         # HTTP controllers (Auth & Tasks)
│   ├── middleware/       # JWT Auth & CORS handling
│   ├── models/           # Data structures (User & Task)
│   └── database/         # DB connection logic
├── web/                  # Frontend assets (HTML, CSS, JS)
├── nginx.conf            # Nginx reverse proxy configuration
├── Dockerfile            # Optimized multi-stage build
└── docker-compose.yml    # Service orchestration
```

## 🚀 Quick Start

### Prerequisites

Before running the project, make sure you have:

-   **Docker**
-   **Docker Compose**
-   A `.env` file with the `JWT_SECRET` variable defined

### Installation

1.  **Clone the repository**

``` bash
git clone https://github.com/your-username/todo-app.git
cd todo-app
```

2.  **Start the environment**

``` bash
docker-compose up --build -d
```

3.  **Access the application**

Open your browser and go to:

    http://localhost

## 🛡 API Endpoints

  -------------------------------------------------------------------------
  Method   Endpoint          Description                    Auth Required
  -------- ----------------- ------------------------------ ---------------
  POST     /auth/signup      Register a new user            No

  POST     /auth/login       Authenticate and get token     No

  GET      /tasks            Fetch all user tasks           Yes (JWT)

  POST     /tasks/create     Create a new task              Yes (JWT)

  POST     /tasks/start      Mark task as In Progress       Yes (JWT)

  POST     /tasks/done       Mark task as Completed         Yes (JWT)

  DELETE   /tasks/delete     Remove a task                  Yes (JWT)
  -------------------------------------------------------------------------

## 🔧 DevOps & CI/CD

The project includes a GitHub Actions workflow in:

    .github/workflows/deploy.yml

This workflow automatically:

-   Connects to a self-hosted runner
-   Builds optimized Docker images
-   Deploys containers with `--force-recreate` to ensure smooth updates

## ✅ Features

-   User registration and login
-   JWT-based authentication
-   Password hashing with bcrypt
-   Task creation, progress tracking, completion, and deletion
-   PostgreSQL persistence
-   Dockerized development and deployment
-   Reverse proxy setup with Nginx
-   Automated deployment with GitHub Actions

## 📌 Notes

-   Make sure your `.env` file is present before starting the
    application.
-   Update the repository URL in the clone command to match your actual
    GitHub repository.
-   For production use, ensure secure secret management and HTTPS
    configuration.

## 📄 License

This project is licensed under the MIT License.
