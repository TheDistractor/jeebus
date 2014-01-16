#include <JeeLib.h>

BlinkPlug bp (1);
uint16_t count;

void setup () {
  Serial.begin(57600);
  Serial.println("\n[blinker]");
}

void loop () {
  Serial.println(++count);
  delay(1000);
}
