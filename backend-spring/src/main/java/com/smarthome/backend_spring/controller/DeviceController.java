package com.smarthome.backend_spring.controller;

import com.smarthome.backend_spring.dto.DeviceDTO;
import com.smarthome.backend_spring.model.Device;
import com.smarthome.backend_spring.model.User;
import com.smarthome.backend_spring.repository.UserRepository;
import com.smarthome.backend_spring.service.DeviceService;
import lombok.RequiredArgsConstructor;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UsernameNotFoundException;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/devices")
@RequiredArgsConstructor
public class DeviceController {

    private final DeviceService deviceService;
    private final UserRepository userRepository;

    
    @PostMapping
    public DeviceDTO createDevice(@RequestBody DeviceDTO dto) {
        
        String currentUsername = SecurityContextHolder.getContext().getAuthentication().getName();
        
        
        User user = userRepository.findByUsername(currentUsername)
                .orElseThrow(() -> new UsernameNotFoundException("User not found with username: " + currentUsername));

        
        Device device = new Device();
        device.setName(dto.getName());
        device.setType(dto.getType());
        device.setStatus(dto.getStatus());

        
        Device savedDevice = deviceService.createDeviceForUser(device, user.getId());
        
        return new DeviceDTO(
            savedDevice.getId(), 
            savedDevice.getName(), 
            savedDevice.getType(), 
            savedDevice.getStatus(), 
            savedDevice.getUserId()
        );
    }

    
    @GetMapping
    public List<DeviceDTO> getAllDevices() {
        
        String currentUsername = SecurityContextHolder.getContext().getAuthentication().getName();
        
        
        User user = userRepository.findByUsername(currentUsername)
                .orElseThrow(() -> new UsernameNotFoundException("User not found with username: " + currentUsername));

        
        return deviceService.getDevicesByUserId(user.getId()).stream()
            .map(device -> new DeviceDTO(
                device.getId(),
                device.getName(),
                device.getType(),
                device.getStatus(),
                device.getUserId()
            ))
            .collect(Collectors.toList());
    }
}