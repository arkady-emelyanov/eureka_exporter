FROM maven:3.5-slim as builder
WORKDIR /app
COPY pom.xml .
RUN mvn verify clean --fail-never
COPY . .
RUN mvn package

## target
FROM azul/zulu-openjdk:8
COPY --from=builder /app/target/auto-service-with-dependencies.jar /auto-service.jar
CMD [ "java", "-jar", "/auto-service.jar"]
