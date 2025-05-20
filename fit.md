# Fit Protocl notes, full details can be seen in: https://developer.garmin.com/fit/protocol/

## Table of Contents

- [Overview](#overview)
- [FIT Profiles](#fit-profiles)
  - [Global Profile](#global-profile)
  - [Product Profile](#product-profile)
- [FIT File Protocol](#fit-file-protocol)
- [FIT File Structure](#fit-file-structure)
  - [File Header](#file-header)
  - [CRC](#crc)
  - [Data Records](#data-records)
- [Record Format](#record-format)
  - [Record Header Byte](#record-header-byte)
  - [Normal Header](#normal-header)
  - [Compressed Timestamp Header](#compressed-timestamp-header)
- [Record Content](#record-content)
  - [Definition Message](#definition-message)
  - [Data Message](#data-message)
- [Scale/Offset](#scaleoffset)
- [Dynamic Fields](#dynamic-fields)
- [Components](#components)
- [Common Fields](#common-fields)
- [Best Practices](#best-practices)
- [Defining Data Messages](#defining-data-messages)
- [Redefining Local Message Types](#redefining-local-message-types)
- [FIT Message Conversion](#fit-message-conversion)
- [Compatibility](#compatibility)
- [Common FIT File Applications](#common-fit-file-applications)
- [Plugin Framework](#plugin-framework)

---

## Overview

A FIT file contains a series of records, each with a header and content. The content is either a **definition message** (specifies upcoming data) or a **data message** (contains data fields). The protocol defines message types, field formats, and data compression methods.

---

## FIT Profiles

### Global Profile

- Maintained by Garmin International, Inc.
- Contains all system configurations, messages, fields, and data types.
- **System Configurations:** Byte endianness, alignment, etc.
- **FIT Messages:** Define fields within each message.
- **FIT Message Fields:** Define base type and format.
- **FIT Types:** Describe field variable types (e.g., unsigned char, signed short).

### Product Profile

- Application-specific subset of the Global Profile.
- Defines only necessary data messages for a product.
- Custom messages can be defined in the manufacturer-specific range (`0xFF00–0xFFFE`).
- Devices ignore unrecognized messages and fill missing data with invalid values.

---

## FIT File Protocol

- Describes how profiles are implemented and files are transferred.
- Data is encoded according to the product profile and transferred between devices.
- Decoding uses the receiving device's product profile.

---

## FIT File Structure

All FIT files have the same structure:

1. **File Header**
2. **Data Records** (encoded FIT messages)
3. **2-byte CRC**

### File Header

- Minimum size: 12 bytes (legacy), 14 bytes preferred.
- Contains protocol/profile version, data size, and ".FIT" signature.
- CRC is optional in 14-byte header.

**Header Format:**

| Byte | Parameter            | Size (Bytes) | Description                                   |
|------|----------------------|--------------|-----------------------------------------------|
| 0    | Header Size          | 1            | Length of header                              |
| 1    | Protocol Version     | 1            | Protocol version number                       |
| 2-3  | Profile Version      | 2            | Profile version number                        |
| 4-7  | Data Size            | 4            | Length of Data Records section (excl. header) |
| 8-11 | Data Type            | 4            | ASCII ".FIT"                                  |
| 12-13| CRC (optional)       | 2            | CRC of bytes 0-11 or 0x0000                   |

### CRC

- Final 2 bytes: 16-bit CRC (little endian).
- CRC calculation:

```c
FIT_UINT16 FitCRC_Get16(FIT_UINT16 crc, FIT_UINT8 byte)
{
   static const FIT_UINT16 crc_table[16] = {
      0x0000, 0xCC01, 0xD801, 0x1400, 0xF001, 0x3C00, 0x2800, 0xE401,
      0xA001, 0x6C00, 0x7800, 0xB401, 0x5000, 0x9C01, 0x8801, 0x4400
   };
   FIT_UINT16 tmp;
   tmp = crc_table[crc & 0xF];
   crc = (crc >> 4) & 0x0FFF;
   crc = crc ^ tmp ^ crc_table[byte & 0xF];
   tmp = crc_table[crc & 0xF];
   crc = (crc >> 4) & 0x0FFF;
   crc = crc ^ tmp ^ crc_table[(byte >> 4) & 0xF];
   return crc;
}
```

---

## Data Records

- Main content of the FIT file.
- Two kinds:
  - **Definition Messages:** Define upcoming data messages.
  - **Data Messages:** Contain data fields as per definition message.
    - Normal Data Message
    - Compressed Timestamp Data Message

Each record starts with a 1-byte **Record Header**.

---

## Record Format

### Record Header Byte

- 1 byte bit field.
- Two types: **Normal Header** and **Compressed Timestamp Header**.

#### Normal Header

- Bit 7: 0 (Normal Header)
- Bit 6: Message Type (1 = Definition, 0 = Data)
- Bit 5: Message Type Specific (see below)
- Bit 4: Reserved
- Bits 0-3: Local Message Type (0-15)

#### Compressed Timestamp Header

- Bit 7: 1 (Compressed Timestamp Header)
- Bits 5-6: Local Message Type (0-3)
- Bits 0-4: Time Offset (seconds, 5 bits)

---

## Record Content

### Definition Message

- Associates local message type with a Global Message Number.
- Format:

| Byte         | Description                | Length      | Value/Notes                                 |
|--------------|---------------------------|-------------|---------------------------------------------|
| 0            | Reserved                  | 1           | 0                                           |
| 1            | Architecture              | 1           | 0: Little Endian, 1: Big Endian             |
| 2-3          | Global Message Number     | 2           | Endianness as per Architecture              |
| 4            | Fields                    | 1           | Number of fields in Data Message            |
| 5+           | Field Definitions         | 3 per field | See below                                   |
| ...          | # Developer Fields        | 1           | If Developer Data Flag set                  |
| ...          | Developer Field Definitions| 3 per field | If Developer Data Flag set                  |

#### Field Definition

| Byte | Name                   | Description                                      |
|------|------------------------|--------------------------------------------------|
| 0    | Field Definition Number| Defined in Global FIT profile                    |
| 1    | Size                   | Size (bytes)                                     |
| 2    | Base Type              | FIT variable type (see Table 7 below)            |

#### Base Types (Table 7)

| #  | Endian | Field   | Type Name | Invalid Value         | Size | Comment                  |
|----|--------|---------|-----------|-----------------------|------|--------------------------|
| 0  | 0      | 0x00    | enum      | 0xFF                  | 1    |                          |
| 1  | 0      | 0x01    | sint8     | 0x7F                  | 1    | 2’s complement           |
| 2  | 0      | 0x02    | uint8     | 0xFF                  | 1    |                          |
| 3  | 1      | 0x83    | sint16    | 0x7FFF                | 2    | 2’s complement           |
| 4  | 1      | 0x84    | uint16    | 0xFFFF                | 2    |                          |
| 5  | 1      | 0x85    | sint32    | 0x7FFFFFFF            | 4    | 2’s complement           |
| 6  | 1      | 0x86    | uint32    | 0xFFFFFFFF            | 4    |                          |
| 7  | 0      | 0x07    | string    | 0x00                  | 1    | Null-terminated UTF-8    |
| 8  | 1      | 0x88    | float32   | 0xFFFFFFFF            | 4    |                          |
| 9  | 1      | 0x89    | float64   | 0xFFFFFFFFFFFFFFFF    | 8    |                          |
| 10 | 0      | 0x0A    | uint8z    | 0x00                  | 1    |                          |
| 11 | 1      | 0x8B    | uint16z   | 0x0000                | 2    |                          |
| 12 | 1      | 0x8C    | uint32z   | 0x00000000            | 4    |                          |
| 13 | 0      | 0x0D    | byte      | 0xFF                  | 1    | Array of bytes           |
| 14 | 1      | 0x8E    | sint64    | 0x7FFFFFFFFFFFFFFF    | 8    | 2’s complement           |
| 15 | 1      | 0x8F    | uint64    | 0xFFFFFFFFFFFFFFFF    | 8    |                          |
| 16 | 1      | 0x90    | uint64z   | 0x0000000000000000    | 8    |                          |

### Developer Data Fields

- Allow custom fields without changing the FIT profile.
- Defined by special global messages (developer_data_id, field_description).

---

## Data Message

- Follows the format specified by the definition message of matching local message type.
- Can be very compact.

---

## Scale/Offset

- Some fields use scale/offset for efficient representation.
- Value = (binary quantity / scale) - offset

---

## Dynamic Fields

- Field interpretation depends on another field's value (subfields).
- Subfields have no field number; reference field/value combinations determine interpretation.

---

## Components

- Compress multiple fields into a bit field in a single containing field.
- Decoder expands components into destination fields.

---

## Common Fields

- **Message Index (Field #254):** Used for indexing messages.
- **Timestamp (Field #253):** UTC timestamp.
- **Part Index (Field #250):** Sequence number for multi-part data.

---

## Best Practices

- Always include:
  - FIT File header
  - file_id Definition Message
  - file_id Data Message
  - Appropriate definition messages before data messages
  - 2-byte CRC

---

## Defining Data Messages

- Data messages must be specified by a definition message first.
- Use a single definition message for multiple data messages of the same format.

---

## Redefining Local Message Types

- Local message types can be redefined within a single FIT file.
- Minimize the number of local message types to save RAM.

---

## FIT Message Conversion

- SDK provides code for conversion between device architectures and protocol versions.
- Unrecognized data is ignored; missing data is filled with invalid values.

---

## Compatibility

- FIT protocol is extensible and backward compatible.
- Endian architecture is handled automatically.

---

## Common FIT File Applications

| FIT File Type | Purpose                                      |
|---------------|----------------------------------------------|
| Settings      | User parameters (Age, Weight, Height)        |
| Activity      | Data/events from an active session           |
| Workout       | Workout parameters (target rates, durations) |
| Blood Pressure| Summary data from blood pressure device      |
| Weight        | Summary data from weight scale device        |

---

## Plugin Framework

- Allows manipulation of FIT files before output to the end-application.
- Implemented in C++, C#, and Java.
- Example: Heart Rate to Record Plugin, Three D Sensor Adjustment Plugin.

---

*For detailed examples, refer to the original document's figures and tables.*
```
This reformats the original markdown for clarity, structure, and readability, using headings, tables, and lists. For further improvements, consider splitting into multiple files or adding diagrams as images.
