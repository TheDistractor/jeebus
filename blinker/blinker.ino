#include <JeeLib.h>

MilliTimer timer;
BlinkPlug bp (1);
uint16_t count;

void setup () {
  Serial.begin(57600);
  Serial.println("\n[blinker]");
}

void loop () {
  if (Serial.read() == 'L') {
    while (!Serial.available())
      ;
    byte led = Serial.read() - '0';
    while (!Serial.available())
      ;
    switch (Serial.read()) {
      case '1': bp.ledOn(led); break;
      case '0': bp.ledOff(led); break;
    }
  }

  switch (bp.buttonCheck()) {
    case BlinkPlug::ON1:  Serial.println("G1"); break;
    case BlinkPlug::OFF1: Serial.println("G0"); break;
    case BlinkPlug::ON2:  Serial.println("R1"); break;
    case BlinkPlug::OFF2: Serial.println("R0"); break;
  }

  if (timer.poll(1000)) {
    Serial.print('C');
    Serial.println(++count);
  }
}
