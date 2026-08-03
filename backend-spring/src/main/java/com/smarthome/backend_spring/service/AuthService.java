package com.smarthome.backend_spring.service;

import com.smarthome.backend_spring.config.JwtCore;
import com.smarthome.backend_spring.dto.AuthRequest;
import com.smarthome.backend_spring.dto.AuthResponse;
import com.smarthome.backend_spring.model.User;
import com.smarthome.backend_spring.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class AuthService {

    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;
    private final JwtCore jwtCore;

   
    public String register(AuthRequest request) {
        if (userRepository.findByUsername(request.getUsername()).isPresent()) {
            throw new RuntimeException("Пользователь с таким именем уже существует!");
        }

        User user = new User();
        user.setUsername(request.getUsername());
       
        user.setPassword(passwordEncoder.encode(request.getPassword()));
        
        userRepository.save(user);
        return "Пользователь успешно зарегистрирован";
    }

   
    public AuthResponse login(AuthRequest request) {
        User user = userRepository.findByUsername(request.getUsername())
                .orElseThrow(() -> new RuntimeException("Пользователь не найден"));

        
        if (!passwordEncoder.matches(request.getPassword(), user.getPassword())) {
            throw new RuntimeException("Неверный пароль");
        }

        
        String token = jwtCore.generateToken(user.getUsername());
        return new AuthResponse(token);
    }
}