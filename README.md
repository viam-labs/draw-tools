# Module draw-motion-plan

A Viam module that visualizes poses as arrows in the World State Store service.

## Model viam-viz:draw-motion-plan:as-arrows

This module integrates with Viam's World State Store service to visualize poses as arrows. It provides a simple interface for drawing custom arrows with specified poses, colors, and reference frames.

### Configuration

The following attribute template can be used to configure this model:

```json
{}
```

#### Attributes

This model currently has no required configuration attributes. The configuration is empty, making it simple to set up.

### DoCommand

The module supports the following commands:

#### Draw

Adds arrows representing poses to the world state. Each arrow can have a custom pose, color, and parent frame.

**Parameters:**

- `draw` (required): Array of arrow objects to draw

Each arrow object in the array should contain:

- `pose` (required): Object containing position and orientation
- `color` (optional): Object containing RGB color values
- `parent_frame` (optional): Reference frame name (defaults to "world")

**Examples:**

Draw a single red arrow at the origin:

```json
{
  "draw": [
    {
      "pose": {
        "x": 0.0,
        "y": 0.0,
        "z": 0.0,
        "o_x": 0.0,
        "o_y": 0.0,
        "o_z": 1.0,
        "theta": 0.0
      },
      "color": {
        "r": 255,
        "g": 0,
        "b": 0
      }
    }
  ]
}
```

Draw multiple arrows with different colors and positions:

```json
{
  "draw": [
    {
      "pose": {
        "x": 1.0,
        "y": 0.0,
        "z": 0.0,
        "o_x": 0.0,
        "o_y": 0.0,
        "o_z": 1.0,
        "theta": 0.0
      },
      "color": {
        "r": 255,
        "g": 0,
        "b": 0
      },
      "parent_frame": "base_link"
    },
    {
      "pose": {
        "x": 0.0,
        "y": 1.0,
        "z": 0.0,
        "o_x": 0.0,
        "o_y": 0.0,
        "o_z": 1.0,
        "theta": 90.0
      },
      "color": {
        "r": 0,
        "g": 255,
        "b": 0
      }
    }
  ]
}
```

**Response:**

```json
{
  "success": true,
  "arrows_added": 2
}
```

#### Clear Arrows

Removes all arrows from the world state.

```json
{
  "clear": {}
}
```

**Response:**

```json
{
  "success": true,
  "arrows_removed": 2
}
```
