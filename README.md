# Social Media API

This is a simple **social media API** built with **Go Fiber**, **GORM**, and **SQLite**. It allows users to register, log in, create posts, like posts, and comment on them. It also includes **Google OAuth SSO** and auto-generates usernames upon registration.

---

## **Features**

- User Authentication (JWT-based login & Google OAuth SSO)
- User Profiles (Profile Picture, Name, Username, Bio, Followers & Following Count)
- CRUD operations for posts (Create, Read, Update, Delete)
- Likes and Comments on posts
- Timeline to fetch all posts
- API Documentation with Swagger

---

## **Installation & Setup**

### **1. Clone the Repository**

```sh
git clone https://github.com/yourusername/socialmedia-api.git
cd socialmedia-api
```

### **2. Install Dependencies**

```sh
go mod tidy
```

### **3. Configure Environment Variables**

Create a **.env** file in the project root and add:

```ini
PORT=8080
DATABASE_URL=./socialmedia.db
JWT_SECRET=your_secret_key

GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

### **4. Run Database Migrations**

```sh
go run main.go
```

This will create the **SQLite database** and apply migrations.

---

## **Running the API**

### **Start the Server**

```sh
go run main.go
```

Server runs on **http://localhost:8080/**.

---

## **API Documentation**

Swagger API docs are available at:

```
http://localhost:8080/swagger/index.html
```

To generate the docs:

```sh
swag init
```

---

## **API Endpoints**

### **Authentication**

- `POST /api/register` → Register a new user
- `POST /api/login` → Login and get JWT token
- `GET /auth/google` → Google OAuth login

### **Users**

- `GET /api/user/:id` → Get user profile
- `POST /api/user/follow/:id` → Follow a user
- `POST /api/user/unfollow/:id` → Unfollow a user

### **Posts**

- `POST /api/posts` → Create a post
- `GET /api/posts/:id` → Get a single post
- `GET /api/timeline` → Get all posts (timeline)
- `PUT /api/posts/:id` → Update a post
- `DELETE /api/posts/:id` → Delete a post

### **Likes & Comments**

- `POST /api/posts/:id/like` → Like a post
- `POST /api/posts/:id/comment` → Comment on a post
- `GET /api/posts/:id/comments` → Get comments on a post

---

## **Testing with cURL**

### **Register a User**

```sh
curl -X POST http://localhost:8080/api/register -H "Content-Type: application/json" -d '{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}'
```

### **Login**

```sh
curl -X POST http://localhost:8080/api/login -H "Content-Type: application/json" -d '{
  "email": "john@example.com",
  "password": "password123"
}'
```

### **Create a Post**

```sh
curl -X POST http://localhost:8080/api/posts -H "Authorization: Bearer <your_token>" -H "Content-Type: application/json" -d '{
  "content": "This is my first post!"
}'
```

---

## **Deployment**

### **Build the API**

```sh
go build -o socialmedia
```

### **Run the Built Application**

```sh
./socialmedia
```

### **Run with Docker**

Create a `Dockerfile`:

```dockerfile
FROM golang:1.19
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .
EXPOSE 8080
CMD ["./main"]
```

Then, build and run:

```sh
docker build -t socialmedia-api .
docker run -p 8080:8080 socialmedia-api
```

---

## **License**

This project is licensed under the MIT License.
