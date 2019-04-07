package com.epam.subbotnik.moto.domain;

import lombok.Builder;
import lombok.Data;

@Builder
@Data
public class Bike {
    private String color;
    private String make;
    private String model;
    private int year;
}
