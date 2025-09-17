# Book Shop API

This is a simple bookshop REST API that allows users to browse books, put them in the cart, and purchase them (without actual payment processing).

## Task scope and expectations

There are 3 permission levels:

- Anonymous users: they should be able to list and filter books by category. It should also be possible to filter by multiple categories, which would return books from all of the categories selected.
- Authenticated users: they should be able to do everything anonymous users can, but can also add books to the cart. After that, they should be able to check out and buy books.
- Administrators: they should be able to CRUD categories and books.

## Functional requirements
- It should be possible for users to register and authenticate using email+password through the API.
- Users can be made admins only through the DB.
- Admins can CRUD categories. Every category has a name and books assigned to it. Categories hierarchy is flat - meaning that they can’t be nested.
- Admins can CRUD books. Every book has a title, year published, author name, price in USD, and category. Each book can (and must) be assigned to a category. Every book also has a number of copies in stock. Books that are sold out should not be visible in the listing, and it should not be possible to buy them. Stock can be specified when a book is created, but can’t be edited later.
- Visitors (including unauthenticated ones) should be able to browse and filter books.
- Authenticated users should be able to add books to their cart. For simplicity, let’s assume that users can buy multiple books, but only one copy of each (so quantity is not necessary).
- There should be an endpoint that completes checkout and “buys” books currently in the cart. Please note that for simplicity, this endpoint should not take in any credit card details. It should simply pretend that it received them and can assume that a payment was made successfully, and should simply clear the cart and reduce the available quantity of books bought.
- Be mindful of race condition cases where two users might want to buy the same book when there is only one in stock. This should not be possible.
- Note that if a user adds a book to the cart and doesn’t buy it, after 30 minutes it needs to be “released” from the user's cart and made available again for the others.

## Technical requirements
- The server needs to expose a GraphQL or RESTful API. It needs to be able to support non-web clients such as mobile apps.
- Be mindful of the edge cases and unexpected scenarios.
- Be mindful of security and data validation.
- It is expected that this system can handle 10 000+ books being offered.

## How to run
- `make dc` runs docker-compose with the app container on port 8080 for you.
- `make test` runs the tests
- `make run` runs the app locally on port 8080 without docker.
- `make lint` runs the linter

## Solution notes
- :trident: clean architecture (handler->service->repository)
- :book: standard Go project layout (well, more or less :blush:)
- :cd: docker compose + Makefile included
- :card_file_box: PostgreSQL migrations included
- :heavy_check_mark: Postman collection included
- :lock: race conditions are handled by transactions and `SELECT ... FOR UPDATE` in the SQL queries

## API Endpoints

The API provides the following functionality:

- **🌐 General**: Health checks and API info (`/health`, `/`)
- **👤 Authentication**: User registration and login (`/signup`, `/signin`)
- **📚 Books**: Browse books and get details (`/books`, `/book/{book_id}`)
- **🏷️ Categories**: Browse categories (`/categories`, `/category/{category_id}`)
- **🛒 Cart**: Shopping cart management (`/cart`, `/checkout`) (🔐 auth required)
- **📖 Admin Books**: Book CRUD operations (`POST /book`, `PATCH /book/{book_id}`, `DELETE /book/{book_id}`) (👑 admin only)
- **🗂️ Admin Categories**: Category CRUD operations (`POST /category`, `PATCH /category/{category_id}`, `DELETE /category/{category_id}`) (👑 admin only)
- **⚡ gRPC Gateway**: All core API endpoints are also available via HTTP/JSON through gRPC Gateway, mapped from `.proto` definitions in `proto/v1/`:
    - **Books Service (gRPC)**: `POST /v1/book`, `GET /v1/book/{id}`, `PATCH /v1/book/{id}`, `DELETE /v1/book/{id}`, `GET /v1/books`
    - **Cart Service (gRPC)**: `PATCH /v1/cart` (update cart), `POST /v1/cart/checkout` (checkout current cart)
## Testing the API

1. **Import Postman Collection**: Import `postman/Bookshop_API.postman_collection.json`
2. **Start the server**: `make dc` or `make run`
3. **Run requests**: Use Postman to test all endpoints with auto-token management

## Tech Stack
- **Language**: Go 1.25
- **Database**: PostgreSQL 15
- **Migrations**: Goose
- **Testing**: testify, mockery
- **CI/CD**: GitHub Actions
- **Containerization**: Docker Compose
