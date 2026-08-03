package com.smarthome.backend_spring.dto;

import lombok.Data;

@Data
public class AuthRequest {
    private String username;
    private String password;
}