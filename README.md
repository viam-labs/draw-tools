# Module draw-tools

A Viam module to support visualizations.

## Model viam-viz:draw-tools:draw-arrows

This module provides the following resources:

1. **draw-arrows-world-state**: A world state store service that allows arrow visualization
2. **clear-arrows-button**: A button component that clears all arrows when pressed
3. **draw-arrows-button**: A button component that draws predefined arrows when pressed

### Model viam-viz:draw-tools:draw-arrows-world-state

This provides a simple interface for drawing custom arrows with specified poses, colors, and reference frames.

#### Configuration

The service does not have any required attributes for configuration, but can accept an `arrows` field to render arrows
when the service is created.

- `arrows` (optional): Array of arrow objects to draw when the service starts. Each arrow object contains:
  - `pose` (required): Object containing position and orientation
  - `name` (optional): Name of the arrow frame (defaults to "arrow-{uuid}")
  - `color` (optional): Object containing RGB color values (defaults to yellow)
  - `parent_frame` (optional): Reference frame name (defaults to "world")
  - `uuid` (optional): UUID string for the arrow (generates new UUID if not provided)

**Configuration**

```json
{
  "arrows": [
    {
      "name": "arrow-1",
      "pose": {
        "x": 1,
        "y": 0,
        "z": 0,
        "o_x": 1,
        "o_y": 0,
        "o_z": 0,
        "theta": 0
      },
      "color": {
        "r": 255,
        "g": 0,
        "b": 0
      }
    },
    {
      "name": "arrow-2",
      "pose": {
        "x": 0,
        "y": 1,
        "z": 0,
        "o_x": 0,
        "o_y": 1,
        "o_z": 0,
        "theta": 90
      },
      "color": {
        "r": 0,
        "g": 255,
        "b": 0
      }
    },
    {
      "name": "arrow-3",
      "pose": {
        "x": 0,
        "y": 0,
        "z": 1,
        "o_x": 0,
        "o_y": 0,
        "o_z": 1,
        "theta": 180
      },
      "color": {
        "r": 0,
        "g": 0,
        "b": 255
      }
    }
  ]
}
```

#### DoCommand

The service supports the following commands:

##### Draw

Adds arrows representing poses to the world state. Each arrow can have a custom pose, color, and parent frame.

**Parameters:**

- `draw` (required): Array of arrow objects to draw

Each arrow object in the array should contain:

- `pose` (required): Object containing position and orientation
- `name` (optional): Name of the arrow frame (defaults to "arrow-{uuid}")
- `color` (optional): Object containing RGB color values (defaults to yellow)
- `parent_frame` (optional): Reference frame name (defaults to "world")
- `uuid` (optional): UUID string for the arrow (generates new UUID if not provided)

**Command:**

```json
{
  "draw": [
    {
      "name": "arrow-4",
      "pose": {
        "x": -1,
        "y": 0,
        "z": 0,
        "o_x": -1,
        "o_y": 0,
        "o_z": 0,
        "theta": 0
      },
      "color": {
        "r": 255,
        "g": 255,
        "b": 0
      }
    },
    {
      "name": "arrow-5",
      "pose": {
        "x": 0,
        "y": -1,
        "z": 0,
        "o_x": 0,
        "o_y": -1,
        "o_z": 0,
        "theta": -90
      },
      "color": {
        "r": 0,
        "g": 255,
        "b": 255
      }
    },
    {
      "name": "arrow-6",
      "pose": {
        "x": 0,
        "y": 0,
        "z": -1,
        "o_x": 0,
        "o_y": 0,
        "o_z": -1,
        "theta": -180
      },
      "color": {
        "r": 255,
        "g": 0,
        "b": 255
      }
    }
  ]
}
```

**Response:**

```json
{
  "success": true,
  "arrows_added": 3
}
```

##### Clear

Removes all arrows from the world state.

**Command:**

```json
{
  "clear": {}
}
```

**Response:**

```json
{
  "success": true,
  "arrows_removed": 3
}
```

### Model viam-viz:draw-tools:clear-arrows-button

A button component that removes all arrows from the world state when pressed. This component connects to an `draw-arrows-world-state` service and triggers the clear command when the button is pushed.

#### Configuration

The following attribute template can be used to configure this component:

```json
{
  "service_name": "draw-arrows-service" // must be included in `depends_on`
}
```

**NOTE**: The `draw-arrows-world-state` you want to manage must be included as a dependency on this components `depends_on` configuration.

##### Attributes

- `service_name` (required): The name of the `draw-arrows-world-state` service to connect to

### Model viam-viz:draw-tools:draw-arrows-button

A button component that draws predefined arrows to the world state when pressed. This component connects to a `draw-arrows-world-state` service and triggers the draw command with preconfigured arrows when the button is pushed.

#### Configuration

The following attribute template can be used to configure this component:

```json
{
  "service_name": "draw-arrows-service", // must be included in `depends_on`
  "arrows": [
    {
      "name": "arrow-4",
      "pose": {
        "x": -1,
        "y": 0,
        "z": 0,
        "o_x": -1,
        "o_y": 0,
        "o_z": 0,
        "theta": 0
      },
      "color": {
        "r": 255,
        "g": 255,
        "b": 0
      }
    },
    {
      "name": "arrow-5",
      "pose": {
        "x": 0,
        "y": -1,
        "z": 0,
        "o_x": 0,
        "o_y": -1,
        "o_z": 0,
        "theta": -90
      },
      "color": {
        "r": 0,
        "g": 255,
        "b": 255
      }
    },
    {
      "name": "arrow-6",
      "pose": {
        "x": 0,
        "y": 0,
        "z": -1,
        "o_x": 0,
        "o_y": 0,
        "o_z": -1,
        "theta": -180
      },
      "color": {
        "r": 255,
        "g": 0,
        "b": 255
      }
    }
  ]
}
```

**NOTE**: The `draw-arrows-world-state` you want to manage must be included as a dependency on this components `depends_on` configuration.

##### Attributes

- `service_name` (required): The name of the `draw-arrows-world-state` service to connect to
- `arrows` (required): Array of arrow objects to draw when the button is pressed. Each arrow object contains:
  - `pose` (required): Object containing position and orientation
  - `name` (optional): Name of the arrow frame (defaults to "arrow-{uuid}")
  - `color` (optional): Object containing RGB color values (defaults to yellow)
  - `parent_frame` (optional): Reference frame name (defaults to "world")
  - `uuid` (optional): UUID string for the arrow (generates new UUID if not provided)

## Model viam-viz:draw-tools:draw-mesh

This module provides the following resources:

1. **draw-mesh-world-state**: A world state store service that allows mesh visualization from PLY files
2. **clear-mesh-button**: A button component that clears all meshes when pressed
3. **draw-mesh-button**: A button component that draws a mesh from a specified file path when pressed

### Model viam-viz:draw-tools:draw-mesh-world-state

This provides a simple interface for drawing 3D meshes from PLY files into the world state.

#### Configuration

The service does not have any required attributes for configuration.

#### DoCommand

The service supports the following commands:

##### Draw

Adds a mesh from a PLY file to the world state. The mesh is positioned at the origin (world frame) by default.

**Parameters:**

- `draw` (required): String path to the PLY file to load and display

**Command:**

```json
{
  "draw": "/path/to/mesh.ply"
}
```

**Response:**

```json
{
  "success": true
}
```

##### Clear Meshes

Removes all meshes from the world state.

**Command:**

```json
{
  "clear": {}
}
```

**Response:**

```json
{
  "success": true,
  "mesh_removed": 1
}
```

### Model viam-viz:draw-tools:clear-mesh-button

A button component that removes all meshes from the world state when pressed. This component connects to a `draw-mesh-world-state` service and triggers the clear command when the button is pushed.

#### Configuration

The following attribute template can be used to configure this component:

```json
{
  "service_name": "draw-mesh-service"
}
```

**NOTE**: The `draw-mesh-world-state` service you want to manage must be included as a dependency in this component's `depends_on` configuration.

##### Attributes

- `service_name` (required): The name of the `draw-mesh-world-state` service to connect to

### Model viam-viz:draw-tools:draw-mesh-button

A button component that draws a 3D mesh from a PLY file to the world state when pressed. This component connects to a `draw-mesh-world-state` service and triggers the draw command with a preconfigured file path when the button is pushed.

#### Configuration

The following attribute template can be used to configure this component:

```json
{
  "service_name": "draw-mesh-service",
  "model_path": "/path/to/mesh.ply"
}
```

**NOTE**: The `draw-mesh-world-state` service you want to manage must be included as a dependency in this component's `depends_on` configuration.

##### Attributes

- `service_name` (required): The name of the `draw-mesh-world-state` service to connect to
- `model_path` (required): Path to the PLY file containing the 3D mesh to display
