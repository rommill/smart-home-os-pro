# Stage 1: Build JAR package using Gradle
FROM gradle:8.5-jdk21 AS builder

WORKDIR /app

# Copy project files and build executable JAR
COPY . .
RUN ./gradlew bootJar --no-daemon

# Stage 2: Run application using lightweight OpenJDK 17 Runtime
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

# Copy built JAR from builder stage
COPY --from=builder /app/build/libs/*.jar app.jar

EXPOSE 8082

ENTRYPOINT ["java", "-jar", "app.jar"]