package com.epam.subbotnik.moto.controller;

import com.epam.subbotnik.moto.domain.Bike;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;
import java.util.List;

@RestController
@RequestMapping("/api/bikes")
public class MotoServiceController {

    @GetMapping
    public ResponseEntity<?> getAllBikes() {
        List<Bike> cars = generateBikeData();
        return ResponseEntity.ok(cars);
    }

    private static List<Bike> generateBikeData() {
        List<Bike> cars = new ArrayList<>();
        cars.add(Bike.builder().color("Black").make("Harley-Davidson").model("FXDR 114").year(2015).build());
        cars.add(Bike.builder().color("Red").make("Kawasaki").model("Z900RS").year(2016).build());
        cars.add(Bike.builder().color("Silver").make("BMW").model("G310GS").year(2017).build());
        cars.add(Bike.builder().color("Orange").make("Honda").model("Monkey").year(2019).build());
        cars.add(Bike.builder().color("Silver-Black").make("Suzuki").model("SV650X").year(2018).build());
        return cars;
    }

}
