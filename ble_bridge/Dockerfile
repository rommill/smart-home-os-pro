# Base image with Python 3.11 runtime
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies required for Bluetooth access
RUN apt-get update && apt-get install -y --no-install-recommends \
    bluez \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Install Python requirements
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY . .

CMD ["python", "main.py"]
