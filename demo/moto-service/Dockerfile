FROM maven:3.5-slim as builder
WORKDIR /app
COPY pom.xml .
RUN mvn verify clean --fail-never
COPY . .
RUN mvn package

## target
FROM azul/zulu-openjdk:8
COPY --from=builder /app/target/moto-service-with-dependencies.jar /moto-service.jar
CMD [ "java", "-jar", "/moto-service.jar"]
