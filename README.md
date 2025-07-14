# Wallet App

A comprehensive wallet transaction service built with Go, Gin, PostgreSQL, and Redis. This service provides secure cryptocurrency wallet management with atomic transactions, idempotency protection, and JWT-based authentication.
- **Multi-Currency Support**: BTC, ETH, ADA wallets
- **Atomic Transactions**: All operations are atomic across wallets and transaction entries
- **Idempotency Protection**: Redis-based double-click guard via `X-Idempotency-Key`
- **JWT Authentication**: Secure token-based authentication with Redis caching
- **Transaction History**: Filterable transaction history with pagination
- **User Isolation**: Each user can only access their own wallet data
- **Docker Support**: Complete containerized setup with Docker Compose

## explaining any decisions you made

### For enhancing readibility, `main.go` and `models.go` sacrifice SRP
I believe that homework assignments should be simple and clear, so I chose to sacrifice the Single Responsibility Principle (SRP) in main.go and models.go in favor of readability for a small-scale homework project. The initialization process should ideally be separated into another Go file, with main.go using `import _ folderPath` to inject dependencies. Similarly, controllers should be organized into separate ResourceNameRouter.go files based on API resources. However, I opted for this design to allow interviewers to understand the entire project scope (implementation, middleware, routing) from the entry point. The same design philosophy applies to all data objects - for example, models/models.go aggregates all data structures (though this approach is not ideal for large-scale projects as it reduces maintainability over time with the addition of more pointer receivers and value receiver functions).

### service layer + repository layer -> handler layer
You may notice that the handlers depend on low-level modules like txManager, which raises concerns about SRP and Dependency Inversion Principle (DIP). The handlers handle both service layer and repository layer responsibilities. However, adhering to the design principle of not over-layering homework projects to maintain readability, I chose this design approach. A rigorous layered architecture would include service and repository layers. In a large project scenario, I would structure the service layer to focus on business logic orchestration (e.g., parameter processing → logic computation → persistence operations), while transaction operations (performDeposit, performWithdrawal, performTransfer in this homework) would be moved to the repository layer, with basic persistence operations having an additional DAO layer.

### Simplify the definition of Repository layer
- Due to simple requirements, I eliminated the DAO layer and directly implemented both DAO operations and transaction operations in the repository implementation, which may violate SRP.
- However, having experienced scenarios where different database solutions require migration, abstracting the Repository into an interface remains necessary.

### layer architecture vs. hexagonal architecture
For a crypto wallet that will support more features, the system will undoubtedly become complex in the future. In such cases, I would start with hexagonal architecture from the beginning. However, given the limited current requirements, I adopted a conventional layered architecture. In this layered architecture, you can observe that the repository implementation is simply an abstraction layer for DTO operations, unlike the repository definition in hexagonal architecture, where the repository is an abstraction layer for all domain persistence operations.

### Maintain the atomicity of transaction through `transaction manager`
- In practice, multi-table operations must guarantee atomicity.

### JWT Authentication
Although not required by the specifications, in practice, each user should only be able to access their own account information or execute transactions on their own account. JWT is one of many choices to meet this practical requirement.
- This belongs to the non-functional domain, so I chose to design it at the middleware layer.

### Idempotency Protection
Although not required by the specifications, in practice, to enhance business stability, I habitually add idempotency mechanisms to transaction or ledger-changing requirements to prevent double-clicking.
- This belongs to the non-functional domain, so I chose to design it at the middleware layer.

### Pagination
- Conforms to common practical requirements in applications. Transaction records will certainly number in the hundreds, so I simply added a pagination mechanism.

## how to setup and run your code
### Using Docker Compose
   1. **Clone and navigate to the project**:
      ```bash
      cd wallet-homework
      ```

   2. **Start all services**:
      ```bash
      docker compose up -d  #(or make docker-run)
      ```

   3. **Seed the database** (in a new terminal):
      ```bash
      make seed
      ```

   4. **Login with tested email to get JWT**:
      ```bash
      curl -X POST http://localhost:8080/auth/login \
      -H "Content-Type: application/json" \
      -d '{"email": "user_001@example.com"}'
      ```

   5. **do something**:
      - please refer to the API doc below
      - remember to add JWT token to the header of request

   6. (optional) Import postman collection and set the environment.local
      - please refer to the file in root -- wallet-homework.postman_collection.json
      - after getting wallet info through apis, fill out them into the environments.local and have fun~
   

## highlight how should reviewer view your code
   - cmd/seed contains seed data
   - main.go shows all design intentions for functional/non-functional requirements
      - functional requirement: router
         - deposit
         - withdrawl
         - transfer
         - query
      - non-functional requirement: middleware
         - security through JWT
         - Idempotency

## areas to be improved
   For homework readability, I ignored many SRP principles in the framework

## how long you spent on the test
   1 day design, 3 day development+testing, 1 day docker packaging + README.md

## which features you chose not to do in the submission
   - Sacrificed SRP in many places to increase readability
   - Transactions don't have retry mechanisms - didn't want to increase homework complexity
   - Not friendly to unit test & integration test: The domain is too small and business logic is simple. The design challenges are mainly in persistence operations, so unit tests and integration tests have limited benefits. Therefore, I chose to skip them. However, I still declared a local mock_repository implementation in case integration tests are needed.
   - Didn't consider database protection in microservice architecture: Assumed that all persistence would only be accessed by the current service. If services other than the wallet service share the database, and the overall product services exist in a microservice form, then database operations must be separated into an MQ + service&pod (workers) pattern to control connection pools to comply with database connection limits.
   - JWT security: Usually TLS is used for connection protection, but this project doesn't implement TLS protection. So you'll see sensitive information in the JWT, but I still add an encryption layer to sensitive information.

## API Documentation

### Authentication Header
All protected endpoints require:
```
Authorization: Bearer <jwt_token>
```

### 1. Deposit to Wallet
```bash
curl -X POST http://localhost:8080/wallets/{wallet_id}/deposit \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: unique-key-123" \
  -d '{"amount": "100.50"}'
```

### 2. Withdraw from Wallet
```bash
curl -X POST http://localhost:8080/wallets/{wallet_id}/withdraw \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: unique-key-456" \
  -d '{"amount": "50.25"}'
```

### 3. Transfer Between Wallets
```bash
curl -X POST http://localhost:8080/wallets/{sender_wallet_id}/transfer \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: unique-key-789" \
  -d '{
    "receiver_wallet_id": "receiver-wallet-uuid",
    "amount": "25.75"
  }'
```

### 4. Get Wallet Balance
```bash
curl -X GET http://localhost:8080/wallets/{wallet_id}/balance \
  -H "Authorization: Bearer <token>"
```

### 5. Get User Wallets
```bash
curl -X GET http://localhost:8080/wallets \
  -H "Authorization: Bearer <token>"
```

### 6. Get Transaction History
```bash
curl -X GET "http://localhost:8080/wallets/{wallet_id}/transactions?limit=10&offset=0" \
  -H "Authorization: Bearer <token>"
```

**Query Parameters**:
- `start_date`: Filter transactions from this date (ISO format)
- `end_date`: Filter transactions until this date (ISO format)
- `counterparty_wallet_id`: Filter by specific counterparty
- `limit`: Number of transactions to return (default: all)
- `offset`: Number of transactions to skip (default: 0)
