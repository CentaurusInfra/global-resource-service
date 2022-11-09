#Download base image ubuntu 18.04
FROM ubuntu:18.04

# LABEL about the custom image
LABEL maintainer="carl Xie <q131172019@gmail.com>"
LABEL version="0.1"
LABEL description="This is custom Docker Image for Resource Region Manager Simulator."

# Disable Prompt During Packages Installation
ARG DEBIAN_FRONTEND=noninteractive

# Update Ubuntu Software repository
RUN apt update

# Copy resourceRegionMgrSimulator and define default command for the container
COPY resourceRegionMgrSimulator /usr/local/bin/resourceRegionMgrSimulator
CMD ["/usr/local/bin/resourceRegionMgrSimulator"]

# Expose Port for the Application 
EXPOSE 9119

