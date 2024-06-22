# Use minimal Alpine as the base image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the prebuilt Go executable and configuration files
COPY dist/wios_server_linux /app/myapp
COPY dist/conf /app/conf

# Set the executable permissions (if needed)
RUN chmod +x /app/myapp

# Keep the container running by executing the Go application
CMD ["./myapp"]
