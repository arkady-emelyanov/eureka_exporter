package com.epam.subbotnik.auto.domain;

import lombok.Builder;
import lombok.Data;

@Builder
@Data
public class Car {
    private String color;
    private String make;
    private String model;
    private int year;
}
