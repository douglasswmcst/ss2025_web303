# WEB303: Microservices & Serverless Applications

Welcome to the official GitHub repository for the **WEB303 Microservices & Serverless Applications** module. This repository contains all the necessary resources, including lecture slides, practical instructions, assignment details, and supplementary materials.

---

## 📖 Module Descriptor

### **General Objectives**

This module introduces students to the principles, design patterns, and implementation techniques of microservices and serverless architectures. Students will gain hands-on experience in developing, deploying, and managing distributed systems using modern tools and cloud-native technologies. The module covers the entire lifecycle of microservices and serverless applications, from initial design to production deployment, including inter-service communication, resilience patterns, testing strategies, and observability. By the module's conclusion, students will have a comprehensive understanding of microservices architecture, serverless computing, and the skills to build scalable, maintainable, and efficient distributed systems.

### **Learning Outcomes**

On completion of the module, students will be able to:

1.  Explain the fundamental concepts of microservices and serverless architectures, including their benefits, trade-offs, and appropriate use cases.
2.  Design and implement microservices using gRPC and Protocol Buffers for efficient inter-service communication.
3.  Apply hexagonal architecture principles to create loosely coupled, maintainable microservices.
4.  Implement resilience patterns such as timeout, retry, and circuit breaker to enhance the reliability of distributed systems.
5.  Develop comprehensive testing strategies for microservices, including unit, integration, and end-to-end testing.
6.  Deploy microservices to Kubernetes, implementing various deployment strategies and managing certificates.
7.  Design and implement serverless applications, understanding their event-driven nature and integration patterns.
8.  Implement observability solutions for microservices and serverless applications, including distributed tracing, metrics, and logging.
9.  Evaluate and apply appropriate patterns for data consistency and state management in distributed systems.
10. Analyze and optimize the performance of microservices and serverless applications, addressing common pitfalls and bottlenecks.

---

## 📊 Assessment Approach

Your final grade is determined by Continuous Assessment (CA), which is divided into a 60% theory component and a 40% practical component.

| Assessment Component | Weighting (%) |
| :--- | :--- |
| **A. Practical Work & Report** | 20% |
| **B. Mid-Term Test** | 20% |
| **C. Programming Assignments** | 20% |
| **D. Final Project** | 40% |
| **Total** | **100%** |

---

## 📚 Subject Matter / Syllabus

This module is structured into the following units:

* **Unit I: Introduction to Microservices and Serverless**
    * Monolithic vs. Microservices Architecture
    * Serverless Computing (FaaS)
    * Characteristics of Cloud-Native Applications

* **Unit II: Designing Microservices**
    * Domain-Driven Design (DDD) for Microservices
    * Microservices Patterns (API Gateway, Database per service, Saga)
    * Service Decomposition Strategies

* **Unit III: Inter-service Communication and Resilience**
    * Synchronous Communication (REST vs. gRPC)
    * Asynchronous Communication (Message Queues, Event Streaming)
    * Resilience Patterns (Circuit Breaker, Retry, Timeout)

* **Unit IV: Building Microservices with gRPC**
    * Introduction to gRPC and Protocol Buffers
    * Implementing gRPC Services in Go
    * Hexagonal Architecture for Microservices

* **Unit V: Testing Microservices**
    * Unit Testing, Integration Testing, and End-to-End Testing
    * Mocking, Contract Testing, and Test Containers

* **Unit VI: Deploying Microservices**
    * Containerization with Docker
    * Orchestration with Kubernetes
    * Deployment Strategies (Rolling updates, Blue-Green, Canary)

* **Unit VII: Serverless Architecture and Implementation**
    * Serverless Frameworks and Platforms (AWS Lambda, Azure Functions, etc.)
    * Designing Serverless Applications
    * Serverless Patterns

---

## 💻 Practicals

There will be 6 hands-on practical sessions throughout the semester:

1.  Set up a development environment for microservices using Go, gRPC, and Docker.
2.  Implement a basic microservice using gRPC and Protocol Buffers.
3.  Implement resilience patterns (circuit breaker, retry, timeout) in a microservices environment.
4.  Develop a comprehensive test suite for a microservice, including unit, integration, and end-to-end tests.
5.  Create a serverless function and integrate it with a microservices application.
6.  Design and implement an event-driven microservices architecture with asynchronous communication.

---

## 📖 Reading List

### Essential Reading

* Newman, S. (2019). *Building microservices: Designing Fine-Grained Systems*. O'Reilly Media.
* Bellemare, A. (2020). *Building event-driven microservices: Leveraging Organizational Data at Scale*.
* Babal, H. (2023). *GRPC Microservices in Go*. Simon and Schuster.
* Anderson, E. (2023). *Building serverless applications on Knative: A Guide to Designing and Writing Serverless Cloud Applications*.

### Additional Reading

* Burns, B., Beda, J., Hightower, K., & Evenson, L. (2022). *Kubernetes: up and running: Dive Into the Future of Infrastructure*. O'Reilly Media.

---

## 📂 Repository Structure

This repository is organized as follows:

.
├── lectures/           # Contains all lecture slides and related materials
├── practicals/         # Contains instructions and starter code for practical sessions
│   ├── practical-01/
│   └── ...
├── assignments/        # Details and submission guidelines for programming assignments
└── project/            # Final project requirements and resources