import streamlit as st
from cognite.client import CogniteClient
import json

client = CogniteClient()


@st.cache_data
def get_configs():
    configs = client.extraction_pipelines.config.retrieve("edge-extractor-1")
    return configs


# st.write(get_configs().config)


config = json.loads(get_configs().config)

# Create UI for modifying the Apps configuration
st.title("Camera Apps Configurations")

for i, app in enumerate(config["Apps"]):
    st.subheader(f'App {app["AppName"]}')
    app["InstanceID"] = st.text_input(f"Instance ID {i}", app["InstanceID"])
    app["Configurations"]["DelayBetweenCapture"] = st.number_input(
        f"Delay Between Capture {i}", value=app["Configurations"]["DelayBetweenCapture"]
    )
    app["Configurations"]["CaptureDurationSec"] = st.number_input(
        f"Capture Duration Sec {i}", value=app["Configurations"]["CaptureDurationSec"]
    )
    app["Configurations"]["MaxParallelWorkers"] = st.number_input(
        f"Max Parallel Workers {i}", value=app["Configurations"]["MaxParallelWorkers"]
    )

    st.header("Trigger Topics")
    for j, topic in enumerate(app["Configurations"]["TriggerTopics"]):
        st.subheader(f"Trigger Topic {j}")
        app["Configurations"]["TriggerTopics"][j] = st.text_input(
            f"Trigger Topic {j}", topic
        )
        if st.button(f"Remove Trigger Topic {j}"):
            app["TriggerTopics"].remove(topic)
            st.experimental_rerun()

    with st.form(key=f"add_trigger_topic_form_{i}"):
        new_topic = st.text_input("New Trigger Topic")
        submit_button = st.form_submit_button(label="Add Trigger Topic")
        if submit_button:
            app["Configurations"]["TriggerTopics"].append(new_topic)
            st.experimental_rerun()

    st.header("List Of Target Cameras")
    for j, camera_id in enumerate(app["Configurations"]["ListOfTargetCameras"]):
        st.subheader(f"Target Camera {j}")
        app["Configurations"]["ListOfTargetCameras"][j] = st.number_input(
            f"Target Camera {j}", value=camera_id
        )
        if st.button(f"Remove Target Camera {j}"):
            app["Configurations"]["ListOfTargetCameras"].remove(camera_id)
            st.experimental_rerun()

    with st.form(key=f"add_target_camera_form_{i}"):
        new_camera_id = st.number_input("New Target Camera ID", min_value=1)
        submit_button = st.form_submit_button(label="Add Target Camera")
        if submit_button:
            app["Configurations"]["ListOfTargetCameras"].append(new_camera_id)
            st.experimental_rerun()

# Save the modified configuration
if st.button("Save Configuration"):
    st.write(config)
    st.success("Configuration saved successfully.")
