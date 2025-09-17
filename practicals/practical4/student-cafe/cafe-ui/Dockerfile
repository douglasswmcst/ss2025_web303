# Stage 1: Build the React app
FROM node:18-alpine as build
WORKDIR /app
COPY package.json ./
COPY package-lock.json ./
RUN npm install
COPY . ./
RUN npm run build

# Stage 2: Serve with Nginx
FROM nginx:stable-alpine
COPY --from=build /app/build /usr/share/nginx/html
# Nginx's default port is 80
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
