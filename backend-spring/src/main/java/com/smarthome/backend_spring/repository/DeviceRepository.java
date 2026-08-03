package com.smarthome.backend_spring.repository;

import com.smarthome.backend_spring.model.Device;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface DeviceRepository extends JpaRepository<Device, Long> {
    
    
    List<Device> findByUserId(Long userId);
}