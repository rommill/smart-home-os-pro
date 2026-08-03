package com.smarthome.backend_spring.model;

import jakarta.persistence.*;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "devices")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class Device {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotBlank(message = "Device name cannot be empty")
    @Column(nullable = false)
    private String name;

    @NotBlank(message = "Device type cannot be empty")
    @Column(nullable = false)
    private String type; 

    @NotNull(message = "Connection status must be specified")
    @Column(nullable = false)
    private Boolean status; 

    @Column(name = "user_id")
    private Long userId;
}