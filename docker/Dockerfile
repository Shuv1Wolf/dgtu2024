# Start with the scratch image
FROM scratch

# Set a working directory
# RUN mkdir /app
WORKDIR /app

# Copy the entire project
COPY ./exe ./bin
COPY ./config ./config

# Set environment variables
ENV HTTP_PORT=8083
ENV HTTP_REDIRECT_PORT=8053
ENV REDIRECT_URL=http://89.46.131.17:8053/v1/fit/callback
ENV CLIENT_ID=
ENV CLIENT_SECRET=

# Expose the port
EXPOSE "8083:8083"
EXPOSE "8053:8053"

# Specify the command to run
CMD ["./bin/app"]
