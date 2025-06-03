# First stage: Build the Flutter web app
FROM --platform=linux/$TARGETARCH debian:bullseye-slim AS build

# Set ARG for platform targeting
ARG TARGETARCH

# Accept SERVER_URL as build argument
ARG SERVER_URL=http://localhost:8080
ARG FALLBACK_SERVER_URL=http://localhost:8080
# Accept single-user credentials as build arguments
ARG USER_EMAIL=""
ARG USER_PASSWORD=""

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    git \
    unzip \
    xz-utils \
    zip \
    libglu1-mesa \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set up Flutter
RUN git clone https://github.com/flutter/flutter.git -b stable /flutter
ENV PATH="/flutter/bin:$PATH"

# Verify flutter installation
RUN flutter doctor

# Set working directory
WORKDIR /app

# Copy the Flutter project files
COPY . .

# Configure Flutter for web
RUN flutter config --enable-web

# Get Flutter packages
RUN flutter pub get

# Initialize web project (creates index.html if missing)
RUN flutter create --platforms=web .

# Build the web app with SERVER_URL and credentials passed as dart-define
RUN flutter build web --release \
    --dart-define=SERVER_URL=${SERVER_URL} \
    --dart-define=FALLBACK_SERVER_URL=${FALLBACK_SERVER_URL} \
    --dart-define=USER_EMAIL=${USER_EMAIL} \
    --dart-define=USER_PASSWORD=${USER_PASSWORD}

# Second stage: Serve the app with Nginx
FROM --platform=linux/$TARGETARCH nginx:alpine

# Copy the build output to replace the default nginx contents
COPY --from=build /app/build/web /usr/share/nginx/html

# Copy custom nginx configuration
COPY --from=build /app/nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start Nginx server
CMD ["nginx", "-g", "daemon off;"]
