package com.smarthome.backend_spring.service;

import com.smarthome.backend_spring.model.Device;
import com.smarthome.backend_spring.repository.DeviceRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor 
public class DeviceService {

    private final DeviceRepository deviceRepository;

    
    public Device createDeviceForUser(Device device, Long userId) {
        device.setUserId(userId);
        return deviceRepository.save(device);
    }

    public Device createDevice(Device device) {
        return deviceRepository.save(device);
    }

    public List<Device> getAllDevices() {
        return deviceRepository.findAll();
    }

    public List<Device> getDevicesByUserId(Long userId) {
        return deviceRepository.findByUserId(userId);
    }

    public void deleteDevice(Long id) {
        deviceRepository.deleteById(id);
    }
}