package com.smarthome.backend_spring.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class DeviceDTO {
    private Long id;
    private String name;
    private String type;
    private Boolean status;
    private Long userId;
}