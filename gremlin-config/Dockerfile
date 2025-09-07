# Use the official Gremlin TinkerPop server as base
FROM tinkerpop/gremlin-server:latest

# Set working directory
WORKDIR /opt/gremlin-server

# Copy custom configuration files
COPY gremlin-config/ /opt/gremlin-server/conf/
# Expose Gremlin server port
EXPOSE 8182
