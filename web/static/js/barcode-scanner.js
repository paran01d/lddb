// Barcode Scanner functionality using QuaggaJS

class BarcodeScanner {
    constructor() {
        this.isInitialized = false;
        this.isScanning = false;
        this.scannerConfig = {
            inputStream: {
                name: "Live",
                type: "LiveStream",
                target: document.querySelector('#scanner'),
                constraints: {
                    width: { ideal: 640 },
                    height: { ideal: 480 },
                    facingMode: "environment" // Use back camera
                }
            },
            locator: {
                patchSize: "medium",
                halfSample: true
            },
            numOfWorkers: navigator.hardwareConcurrency || 2,
            frequency: 10,
            decoder: {
                readers: [
                    "ean_reader",      // EAN-13 (most common for UPC)
                    "ean_8_reader",    // EAN-8
                    "code_128_reader", // Code 128
                    "code_39_reader",  // Code 39
                    "codabar_reader"   // Codabar
                ]
            },
            locate: true,
            debug: {
                showCanvas: true,
                showPatches: false,
                showFoundPatches: false,
                showSkeleton: false,
                showLabels: false,
                showPatchLabels: false,
                showRemainingPatchLabels: false,
                boxFromPatches: {
                    showTransformed: false,
                    showTransformedBox: false,
                    showBB: false
                }
            }
        };
    }

    // Initialize the barcode scanner
    async init() {
        return new Promise((resolve, reject) => {
            if (this.isInitialized) {
                resolve();
                return;
            }

            // Check if QuaggaJS is loaded
            if (typeof Quagga === 'undefined') {
                reject(new Error('QuaggaJS library not loaded'));
                return;
            }

            // Check for camera permissions
            if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
                reject(new Error('Camera not supported in this browser'));
                return;
            }

            // Request camera permission
            navigator.mediaDevices.getUserMedia({ video: true })
                .then(() => {
                    this.isInitialized = true;
                    resolve();
                })
                .catch(error => {
                    console.error('Camera permission denied:', error);
                    reject(new Error('Camera permission required for barcode scanning'));
                });
        });
    }

    // Start scanning
    async startScanning() {
        if (this.isScanning) {
            console.log('Scanner already running');
            return;
        }

        try {
            await this.init();
            
            return new Promise((resolve, reject) => {
                // Clear any existing handlers
                Quagga.offDetected();
                Quagga.offProcessed();
                
                Quagga.init(this.scannerConfig, (err) => {
                    if (err) {
                        console.error('QuaggaJS initialization failed:', err);
                        reject(new Error('Failed to initialize barcode scanner'));
                        return;
                    }

                    console.log('QuaggaJS initialized successfully');
                    
                    // Set up barcode detection
                    Quagga.onDetected(this.handleBarcodeDetected.bind(this));
                    
                    // Set up visual feedback
                    Quagga.onProcessed((result) => {
                        const drawingCtx = Quagga.canvas.ctx.overlay;
                        const drawingCanvas = Quagga.canvas.dom.overlay;

                        if (result && drawingCtx && drawingCanvas) {
                            // Clear previous drawings
                            drawingCtx.clearRect(0, 0, 
                                parseInt(drawingCanvas.getAttribute("width")), 
                                parseInt(drawingCanvas.getAttribute("height"))
                            );

                            // Draw detection boxes
                            if (result.boxes) {
                                result.boxes.filter(box => box !== result.box).forEach(box => {
                                    Quagga.ImageDebug.drawPath(box, {x: 0, y: 1}, drawingCtx, {
                                        color: "green", 
                                        lineWidth: 2
                                    });
                                });
                            }

                            // Draw the main detection box
                            if (result.box) {
                                Quagga.ImageDebug.drawPath(result.box, {x: 0, y: 1}, drawingCtx, {
                                    color: "#00F", 
                                    lineWidth: 2
                                });
                            }

                            // Draw the barcode line
                            if (result.codeResult && result.codeResult.code) {
                                Quagga.ImageDebug.drawPath(result.line, {x: 'x', y: 'y'}, drawingCtx, {
                                    color: 'red', 
                                    lineWidth: 3
                                });
                            }
                        }
                    });

                    // Start the scanner
                    Quagga.start();
                    
                    // Wait for video to be ready before marking as scanning
                    setTimeout(() => {
                        this.isScanning = true;
                        console.log('Scanner started and ready for detection');
                        resolve();
                    }, 1000);
                });
            });
        } catch (error) {
            console.error('Failed to start scanning:', error);
            throw error;
        }
    }

    // Stop scanning
    stopScanning() {
        if (!this.isScanning) {
            return;
        }

        try {
            // Clean up event handlers
            Quagga.offDetected();
            Quagga.offProcessed();
            
            // Stop the scanner
            Quagga.stop();
            this.isScanning = false;
            console.log('Barcode scanning stopped');
            
            // Clear the scanner display and restore placeholder
            const scannerElement = document.querySelector('#scanner');
            if (scannerElement) {
                scannerElement.innerHTML = `
                    <div class="scanner-placeholder">
                        <p>📷 Camera stopped</p>
                        <p class="scanner-instructions">Click "Start Camera" to begin scanning</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error stopping scanner:', error);
        }
    }

    // Handle barcode detection
    handleBarcodeDetected(result) {
        const code = result.codeResult.code;
        const confidence = result.codeResult.confidence || 0;
        
        console.log('Barcode detected:', code, 'Confidence:', confidence);
        
        // Only accept high-confidence results
        if (confidence > 80) {
            // Stop scanning immediately after detection
            this.stopScanning();
            
            // Show notification
            showNotification(`✅ Barcode detected: ${code}`, 'success');
            
            // Fill in the manual UPC field
            const manualUpcInput = document.getElementById('manual-upc');
            if (manualUpcInput) {
                manualUpcInput.value = code;
            }
            
            // Automatically trigger lookup
            setTimeout(() => {
                if (typeof lookupUPC === 'function') {
                    lookupUPC();
                }
            }, 500);
        } else {
            console.log('Low confidence detection, continuing scan...');
        }
    }

    // Check if the scanner is currently active
    isActive() {
        return this.isScanning;
    }

    // Get camera capabilities
    async getCameraInfo() {
        try {
            const devices = await navigator.mediaDevices.enumerateDevices();
            const videoDevices = devices.filter(device => device.kind === 'videoinput');
            return {
                hasCamera: videoDevices.length > 0,
                cameraCount: videoDevices.length,
                devices: videoDevices.map(device => ({
                    id: device.deviceId,
                    label: device.label || `Camera ${videoDevices.indexOf(device) + 1}`
                }))
            };
        } catch (error) {
            console.error('Error getting camera info:', error);
            return {
                hasCamera: false,
                cameraCount: 0,
                devices: []
            };
        }
    }

    // Switch camera if multiple cameras are available
    async switchCamera(deviceId) {
        if (this.isScanning) {
            this.stopScanning();
        }

        this.scannerConfig.inputStream.constraints.deviceId = deviceId;
        
        try {
            await this.startScanning();
        } catch (error) {
            console.error('Failed to switch camera:', error);
            throw error;
        }
    }
}

// Enhanced Scanner UI Controller
class ScannerUI {
    constructor() {
        this.scanner = new BarcodeScanner();
        this.setupUI();
    }

    setupUI() {
        // Update scanner element content
        const scannerElement = document.getElementById('scanner');
        if (scannerElement) {
            scannerElement.innerHTML = `
                <div class="scanner-placeholder">
                    <p>📷 Camera will appear here</p>
                    <p class="scanner-instructions">Position the LaserDisc barcode in the camera view</p>
                </div>
            `;
        }

        // Add scanner controls
        this.addScannerControls();
    }

    addScannerControls() {
        const modal = document.getElementById('scan-modal');
        if (!modal) return;

        // Add scanner controls after the scanner div
        const scannerDiv = modal.querySelector('#scanner');
        if (!scannerDiv) return;

        const controlsDiv = document.createElement('div');
        controlsDiv.className = 'scanner-controls';
        controlsDiv.innerHTML = `
            <div class="scanner-buttons">
                <button id="start-scan-btn" class="primary-btn">📷 Start Camera</button>
                <button id="stop-scan-btn" class="secondary-btn" style="display: none;">⏹️ Stop Camera</button>
                <button id="torch-btn" class="accent-btn" style="display: none;">🔦 Torch</button>
            </div>
            <div class="scanner-status">
                <p id="scanner-status-text">Ready to scan</p>
            </div>
        `;

        scannerDiv.parentNode.insertBefore(controlsDiv, scannerDiv.nextSibling);

        // Attach event listeners
        this.attachControlListeners();
    }

    attachControlListeners() {
        const startBtn = document.getElementById('start-scan-btn');
        const stopBtn = document.getElementById('stop-scan-btn');
        const torchBtn = document.getElementById('torch-btn');

        if (startBtn) {
            startBtn.addEventListener('click', () => this.startScanning());
        }

        if (stopBtn) {
            stopBtn.addEventListener('click', () => this.stopScanning());
        }

        if (torchBtn) {
            torchBtn.addEventListener('click', () => this.toggleTorch());
        }
    }

    updateStatus(message, type = 'info') {
        const statusElement = document.getElementById('scanner-status-text');
        if (statusElement) {
            statusElement.textContent = message;
            statusElement.className = `scanner-status-${type}`;
        }
    }

    async startScanning() {
        try {
            this.updateStatus('Starting camera...', 'info');
            
            await this.scanner.startScanning();
            
            this.updateStatus('Camera ready - scan a barcode', 'success');
            this.updateButtonStates(true);
            
        } catch (error) {
            console.error('Failed to start scanning:', error);
            this.updateStatus(`Error: ${error.message}`, 'error');
            
            if (error.message.includes('permission')) {
                showNotification('Camera permission is required for barcode scanning. Please allow camera access and try again.', 'error');
            } else if (error.message.includes('not supported')) {
                showNotification('Camera not supported in this browser. Try a different browser or device.', 'error');
            } else {
                showNotification('Failed to start camera. Check if another app is using the camera.', 'error');
            }
        }
    }

    stopScanning() {
        this.scanner.stopScanning();
        this.updateStatus('Camera stopped', 'info');
        this.updateButtonStates(false);
    }

    updateButtonStates(isScanning) {
        const startBtn = document.getElementById('start-scan-btn');
        const stopBtn = document.getElementById('stop-scan-btn');
        const torchBtn = document.getElementById('torch-btn');

        if (startBtn && stopBtn) {
            startBtn.style.display = isScanning ? 'none' : 'inline-block';
            stopBtn.style.display = isScanning ? 'inline-block' : 'none';
        }

        if (torchBtn) {
            torchBtn.style.display = isScanning ? 'inline-block' : 'none';
        }
    }

    async toggleTorch() {
        try {
            const track = this.scanner.getCurrentVideoTrack();
            if (track && 'torch' in track.getCapabilities()) {
                const constraints = track.getConstraints();
                const currentTorch = constraints.advanced && constraints.advanced[0] && constraints.advanced[0].torch;
                
                await track.applyConstraints({
                    advanced: [{ torch: !currentTorch }]
                });
                
                this.updateStatus(currentTorch ? 'Torch off' : 'Torch on', 'info');
            } else {
                showNotification('Torch not supported on this device', 'warning');
            }
        } catch (error) {
            console.error('Failed to toggle torch:', error);
            showNotification('Failed to toggle torch', 'error');
        }
    }
}

// Global scanner instance
let barcodeScanner = null;

// Initialize scanner when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    barcodeScanner = new ScannerUI();
});

// Enhanced modal handling for scanner
function openScanModal() {
    openModal('scan');
    
    // Auto-start camera when modal opens (optional)
    setTimeout(() => {
        if (barcodeScanner) {
            // Uncomment to auto-start camera
            // barcodeScanner.startScanning();
        }
    }, 500);
}

// Enhanced modal closing for scanner
function closeScanModal() {
    if (barcodeScanner) {
        barcodeScanner.stopScanning();
    }
    closeModals();
}

// Override the original modal functions to handle scanner properly
const originalCloseModals = window.closeModals;
window.closeModals = function() {
    if (barcodeScanner && barcodeScanner.scanner.isActive()) {
        barcodeScanner.stopScanning();
    }
    originalCloseModals();
};

// Export scanner functions
window.openScanModal = openScanModal;
window.closeScanModal = closeScanModal;
window.barcodeScanner = barcodeScanner;