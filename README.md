# Module draw-motion-plan

A Viam module that visualizes motion plan poses as arrows in the World State Store service.

## Model viam-viz:draw-motion-plan:as-arrows

This module integrates with Viam's World State Store service to visualize motion plan poses as arrows. It supports both manual and automatic modes for updating the visualization.

### Configuration

The following attribute template can be used to configure this model:

```json
{
  "motion_service": "builtin",
  "update_rate_hz": 10
}
```

#### Attributes

The following attributes are available for this model:

| Name             | Type   | Inclusion | Description                                   |
| ---------------- | ------ | --------- | --------------------------------------------- |
| `motion_service` | string | Required  | The name of the motion service to use         |
| `update_rate_hz` | float  | Optional  | Rate (Hz) to automatically fetch motion plans |

#### Example Configuration

**Manual Mode (No Automatic Updates):**

```json
{
  "motion_service": "builtin"
}
```

**Automatic Mode (Updates at 10 Hz):**

```json
{
  "motion_service": "builtin",
  "update_rate_hz": 10.0
}
```

### DoCommand

The module supports the following commands:

#### Draw Motion Plan

Fetches a motion plan and adds arrows representing each pose to the world state. If no `execution_id` is provided, it automatically uses the most recent motion plan.

**Parameters:**

- `component_name` (optional): Name of the component to get motion plan for. If not provided, uses the most recent plan for any component.
- `execution_id` (optional): Specific execution ID to use. If not provided, automatically uses the most recent execution.

**Examples:**

Get the most recent motion plan for any component:

```json
{
  "draw_motion_plan": {}
}
```

Get the most recent motion plan for a specific component:

```json
{
  "draw_motion_plan": {
    "component_name": "my-arm"
  }
}
```

Get a specific execution's motion plan:

```json
{
  "draw_motion_plan": {
    "component_name": "my-arm",
    "execution_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

**Response:**

```json
{
  "success": true,
  "arrows_added": 15
}
```

#### Clear Arrows

Removes all arrows from the world state.

```json
{
  "clear_arrows": {}
}
```

**Response:**

```json
{
  "success": true,
  "arrows_removed": 15
}
```
