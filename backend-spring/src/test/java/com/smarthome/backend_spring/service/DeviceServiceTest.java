package com.smarthome.backend_spring.service;

import com.smarthome.backend_spring.model.Device;
import com.smarthome.backend_spring.repository.DeviceRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class DeviceServiceTest {

    @Mock
    private DeviceRepository deviceRepository;

    @InjectMocks
    private DeviceService deviceService;

    private Device sampleDevice;

    @BeforeEach
    void setUp() {
        
        sampleDevice = new Device(1L, "Smart Thermostat", "THERMOSTAT", true, 1L);
    }

    @Test
    void shouldSaveDeviceSuccessfully() {
        // Given 
        when(deviceRepository.save(any(Device.class))).thenReturn(sampleDevice);

        // When 
        Device savedDevice = deviceService.createDevice(sampleDevice);

        // Then 
        assertNotNull(savedDevice);
        assertEquals("Smart Thermostat", savedDevice.getName());
        assertEquals("THERMOSTAT", savedDevice.getType());
        assertTrue(savedDevice.getStatus());
        assertEquals(1L, savedDevice.getUserId());

        
        verify(deviceRepository, times(1)).save(sampleDevice);
    }

    @Test
    void shouldReturnAllDevices() {
        // Given
        Device secondDevice = new Device(2L, "Kitchen Light", "LIGHT", false, 1L);
        List<Device> mockDevices = Arrays.asList(sampleDevice, secondDevice);
        
        when(deviceRepository.findAll()).thenReturn(mockDevices);

        // When
        List<Device> result = deviceService.getAllDevices();

        // Then
        assertNotNull(result);
        assertEquals(2, result.size());
        assertEquals("Smart Thermostat", result.get(0).getName());
        assertEquals("Kitchen Light", result.get(1).getName());

        verify(deviceRepository, times(1)).findAll();
    }
}