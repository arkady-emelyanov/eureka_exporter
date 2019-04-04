package com.epam.subbotnik.auto.controller;

import com.epam.subbotnik.auto.domain.Car;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;
import java.util.List;

@RestController
@RequestMapping("/api/cars")
public class AutoServiceController {

    @GetMapping
    public ResponseEntity<?> getAllCars() {
        List<Car> cars = generateCarData();
        return ResponseEntity.ok(cars);
    }

    private static List<Car> generateCarData() {
        List<Car> cars = new ArrayList<>();
        cars.add(Car.builder().color("Red").make("Ferrari").model("F50").year(2015).build());
        cars.add(Car.builder().color("Green").make("Chevrolet").model("Niva").year(2007).build());
        cars.add(Car.builder().color("Black").make("Nissan").model("X-Trail").year(2016).build());
        cars.add(Car.builder().color("White").make("Audi").model("A5").year(2011).build());
        cars.add(Car.builder().color("Green").make("Toyota").model("Land Cruiser Prado").year(2001).build());
        return cars;
    }

}
