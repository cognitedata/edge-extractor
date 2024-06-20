import streamlit as st
from cognite.client import CogniteClient
import json

st.title("Camera manager")
client = CogniteClient()


@st.cache_data
def get_configs():
    configs = client.extraction_pipelines.config.retrieve("edge-extractor-1")
    return configs


# st.write(get_configs().config)


config = json.loads(get_configs().config)

st.header("Camera Configuration")
for i, camera in enumerate(config["Integrations"]["ip_cams_to_cdf"]["Cameras"]):
    st.subheader(f'Camera {camera["ID"]}')
    camera["Name"] = st.text_input(f"Name {i}", camera["Name"])
    camera["Address"] = st.text_input(f"Address {i}", camera["Address"])
    camera["Username"] = st.text_input(f"Username {i}", camera["Username"])
    camera["Password"] = st.text_input(f"Password {i}", camera["Password"])
    camera["EnableCameraEventStream"] = st.checkbox(
        f"Enable Camera Event Stream {i}", camera["EnableCameraEventStream"]
    )
    st.subheader("Camera Event Filters")
    for j, filter in enumerate(camera["EventFilters"]):
        st.subheader(f"Filter {j}")
        filter["TopicFilter"] = st.text_input(
            f"Topic Filter {j}", filter["TopicFilter"]
        )
        filter["ContentFilter"] = st.text_input(
            f"Content Filter {j}", filter["ContentFilter"]
        )
        if st.button(f"Remove Filter {j}"):
            camera["EventFilters"].remove(filter)
            st.experimental_rerun()

    with st.form(key=f"add_filter_form_{i}"):
        new_filter = {
            "TopicFilter": st.text_input("Topic Filter"),
            "ContentFilter": st.text_input("Content Filter"),
        }
        submit_button = st.form_submit_button(label="Add Filter")
        if submit_button:
            camera["EventFilters"].append(new_filter)
            st.experimental_rerun()
    if st.button(f"Remove Camera {i}"):
        config["Integrations"]["ip_cams_to_cdf"]["Cameras"].remove(camera)
        st.experimental_rerun()

st.header("Add New Camera")
with st.form(key="add_camera_form"):
    new_camera = {
        "ID": st.number_input("ID", min_value=1),
        "Name": st.text_input("Name"),
        "Model": st.text_input("Model"),
        "Address": st.text_input("Address"),
        "Username": st.text_input("Username"),
        "Password": st.text_input("Password"),
        "Mode": st.text_input("Mode"),
        "PollingInterval": st.number_input("Polling Interval", min_value=-1, value=-1),
        "State": st.text_input("State"),
        "EnableCameraEventStream": st.checkbox("Enable Camera Event Stream"),
        "EventFilters": [],
    }
    submit_button = st.form_submit_button(label="Add Camera")
    if submit_button:
        config["Integrations"]["ip_cams_to_cdf"]["Cameras"].append(new_camera)
        st.experimental_rerun()

# Save the modified configuration
if st.button("Save Configuration"):
    with open("config_sumitomo.json", "w") as f:
        json.dump(config, f)
    st.success("Configuration saved successfully.")
