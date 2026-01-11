**Role:** Senior Process Mining Consultant & Data Scientist specialising in enterprise business transformation.

**Objective:** Write a `.ipynb` notebook that performs a full end‑to‑end process mining analysis for a specified enterprise process (e.g., order‑to‑cash, IT support tickets).  Use the **PM4Py** framework integrated with Python’s data‑science stack (Pandas, NumPy, Matplotlib).  The notebook should be comprehensive, modular and reproducible, following the tasks and guidelines below.

---

### Notebook Structure and Tasks

1. **Environment Setup & Data Acquisition**

   - Install required packages (pm4py, pandas, numpy, matplotlib) via pip.  Use the notebook’s first code cell to install and import these libraries.

   - Load event logs in both **XES** and **CSV** formats.  For XES, read the log using `pm4py.read_xes` (e.g., `log = pm4py.read_xes("path/to/file.xes")`), then discover a Petri net using the Inductive Miner and visualise it【969964019365071†L84-L91】.  For CSV, ingest the data with pandas and convert it to an event log using `pm4py.objects.conversion.log.converter` after converting timestamp columns【969964019365071†L129-L136】.  Document any assumptions about column names and types.

   - Explain that event logs must include a case identifier, activity name and timestamp; real‑world data often requires extraction and conversion to the IEEE XES standard【378107908900392†L1038-L1043】.

2. **Data Preparation & Cleaning**

   - **Column Mapping:** rename imported CSV columns to PM4Py standard keys: `case:concept:name` (case ID), `concept:name` (activity) and `time:timestamp` (timestamp).  Use `dataframe_utils.convert_timestamp_columns_in_df` to ensure timestamps are converted to datetime【969964019365071†L129-L136】.

   - **Duplicate/Noise Removal:** remove duplicate records and irrelevant data.  The healthcare case study stresses that deleting duplicate records and eliminating noise and outliers is essential for improved outcomes【378107908900392†L1041-L1044】.  Use pandas’ `drop_duplicates` and filtering techniques to clean the data.

   - **Missing Values:** handle missing timestamps and other values through imputation or removal.  For instance, approximate missing time values using the mean of the same event in other cases【378107908900392†L1044-L1049】.  Document the imputation strategy.

   - **Filtering:** apply PM4Py’s filter functions to refine the event log: `filter_start_activities`, `filter_end_activities`, `filter_event_attribute_values`, `filter_trace_attribute_values`, `filter_variants` and `filter_time_range`【378107908900392†L1050-L1072】.  Provide code examples and explain the purpose of each filter.

   - **Privacy Protection:** when dealing with sensitive data, apply differential privacy or anonymization.  Use PM4Py’s anonymization algorithms: control‑flow anonymization via trace‑variant queries (e.g., SaCoFa) with parameters `ε`, `k` and `p`【607127984499128†L88-L124】, and contextual anonymization using the PRIPEL algorithm for timestamps and resources【607127984499128†L126-L134】.  Note that smaller `ε` yields stronger privacy guarantees【607127984499128†L108-L119】.

3. **Exploratory Data Analysis (EDA)**

   - **Basic Statistics:** compute and display the number of events, cases and unique variants.  Variants are unique sequences of activities; some processes may have many variants (e.g., 116 variants in a building‑permit log【608708606862785†L114-L118】).  Use `pm4py.get_variants` to obtain variant statistics.

   - **Activity Distribution:** visualise event counts over time.  Plot distributions by hour, day and month to identify peaks and idle periods; such distributions were used in a hospital case study to identify busy times【378107908900392†L1318-L1330】.

   - **Case Durations and Throughput:** calculate case durations using `pm4py.get_all_case_durations`【969964019365071†L109-L115】.  Compute descriptive statistics (mean, median, quartiles) and visualise the distribution of throughput times using histograms or boxplots.

   - **Variant Analysis:** analyse the frequency of process variants.  Build a Pareto chart to show that a small number of variants often account for most cases—for example, a building‑permit dataset showed that about half of all cases followed a single variant【608708606862785†L122-L124】.  Display the top variants and their percentage of total cases.

   - **Start/End Activities:** use `pm4py.get_start_activities` and `pm4py.get_end_activities` to identify frequent start and end points【969964019365071†L117-L123】.

   - **Additional Statistics:** compute events per case, case arrival rates and average inter‑arrival times【969964019365071†L163-L170】.  Present these metrics in tables or simple charts.

4. **Process Discovery**

   - Implement and compare at least two discovery algorithms:

     * **Inductive Miner:** use `pm4py.discover_petri_net_inductive(log)` to obtain a block‑structured, sound Petri net【969964019365071†L84-L91】.  Set a noise threshold parameter to prune infrequent behaviour.

     * **Heuristic Miner:** apply `pm4py.discover_petri_net_heuristics(log)` or the heuristics miner functions.  This algorithm uses frequency‑based filtering to remove noise; it excludes infrequent paths to create robust models【362502486590578†L240-L253】.  Adjust dependency and frequency thresholds.

   - Visualise the resulting models as Petri nets or Directly‑Follows Graphs (DFGs) using PM4Py’s visualization factories.  Use factory methods to separate import, algorithm execution and visualization【943788946304099†L138-L153】.

   - Provide guidance on algorithm selection: the Inductive Miner guarantees soundness but may be sensitive to noise, whereas the Heuristic Miner handles noisy data and detects short loops but does not guarantee soundness【608708606862785†L154-L180】.  Encourage trying both and comparing results.

5. **Conformance Checking & Model Evaluation**

   - Evaluate discovered models using the five quality dimensions:

     * **Fitness:** assess how much of the log behaviour is captured by the model.  PM4Py computes fitness via token‑based replay and alignments, returning average trace fitness and the proportion of fully fitted traces【378107908900392†L1170-L1182】.

     * **Precision:** measure how much behaviour allowed by the model is not observed in the log.  PM4Py offers ETConformance (token‑based) and Align‑ETConformance (alignment‑based) measures【378107908900392†L1183-L1202】; alignments are more precise but slower.

     * **Generalization:** evaluate the model’s ability to generalise beyond the log.  PM4Py computes generalization by replaying the log and applying a formula based on transition frequencies【378107908900392†L1206-L1225】.

     * **Simplicity:** quantify model complexity; PM4Py measures simplicity as the inverse arc degree of the Petri net【378107908900392†L1263-L1276】.

     * **Soundness:** ensure the model is free of anomalies (dead transitions).  Use WOFLAN to test soundness【378107908900392†L1227-L1231】.

   - Provide code to compute these metrics using PM4Py functions.  Summarise the results in a table comparing different miners; for instance, the healthcare case study found that the Inductive Miner produced perfect fitness but lower precision, while the Heuristic Miner had higher simplicity but lacked soundness【378107908900392†L1278-L1297】.

   - Perform conformance checking using both token‑based replay and alignments; discuss diagnostics such as deviations and missing activities【378107908900392†L1331-L1340】.

6. **Performance Analysis & Organizational Mining**

   - **Bottleneck Detection:** compute sojourn times for each activity to identify long waiting or processing times.  The hospital study presented average sojourn times per activity and used them to pinpoint bottlenecks such as "Recovery Ward" and "Prepare the patient"【378107908900392†L1421-L1449】.  Calculate these using PM4Py or pandas and visualise them with bar charts.

   - **Throughput & Arrival Metrics:** compute throughput time, case arrival ratio and length of stay (LOS) metrics【378107908900392†L1393-L1401】.  Plot event distributions over hours, days and months to reveal busy periods and slack times【378107908900392†L1318-L1330】.

   - **Resource Utilisation & Social Network Analysis:** apply PM4Py’s social network analysis tools to uncover resource relationships, such as handover‑of‑work and working‑together patterns【943788946304099†L193-L204】.  Build handover networks using networkx or Pyvis and identify central or overloaded resources.

   - **Bottleneck Recommendations:** correlate resource workload and sojourn times to recommend interventions, such as adding capacity during peak hours or redistributing tasks.  For example, the hospital study recommended allocating more resources between 7 am and 4 pm when many activities experienced bottlenecks【378107908900392†L1472-L1483】.

7. **Conclusion & Actionable Insights**

   - Summarise key findings from process discovery, conformance checking and performance analysis.  Highlight frequent variants, bottlenecks, deviations and quality metrics results.

   - Provide actionable recommendations for process improvements, such as refining process designs, reducing variants, reallocating resources or updating reference models.  Emphasise practical steps for stakeholders.

8. **Best‑Practice Guidelines for the Notebook**

   - **Modular Design:** follow PM4Py’s architectural guideline of separating objects (event logs and models), algorithms (discovery, conformance, enhancement) and visualisations into distinct modules【943788946304099†L138-L153】.  Encapsulate code into functions or classes for re‑use and clarity.

   - **Frequency Thresholding:** apply frequency thresholds during discovery to filter out infrequent behaviour.  The Heuristic Miner explicitly filters out the most infrequent paths to produce robust models【362502486590578†L240-L253】; similar thresholds should be considered with other miners.

   - **Interpretation Cells:** include a markdown cell after each major code block to explain the purpose of the analysis and how to interpret the results, making the notebook accessible to business users.

   - **Documentation & Versioning:** organise the notebook with clear headings and numbered sections.  Use version identifiers (e.g., R1.00 for requirements, D1.01 for design) and comments to document the code.

   - **Reproducibility:** provide instructions for obtaining or loading the event log.  Set random seeds where appropriate and document assumptions so others can reproduce the analysis.

### References

1. **PM4Py Architecture** – the PM4Py architecture paper recommends strict separation between objects (event logs), algorithms and visualisations to facilitate understanding and reuse【943788946304099†L138-L153】.
2. **Data Cleaning & Filtering** – a healthcare process mining study highlights the necessity of removing duplicate records, eliminating noise and outliers, imputing missing values using mean timestamps and applying filter functions such as `filter_start_activities`, `filter_end_activities`, `filter_event_attribute_values`, `filter_trace_attribute_values`, `filter_variants` and `filter_time_range`【378107908900392†L1041-L1072】.
3. **Privacy‑Preserving Anonymization** – PM4Py integrates differential privacy anonymization techniques: control‑flow anonymization via trace‑variant queries (e.g., SaCoFa) with parameters ε, k and p, and contextual anonymization using the PRIPEL algorithm【607127984499128†L88-L124】【607127984499128†L126-L134】.
4. **Process Variants** – a process mining case study on building permits defines variants as unique sequences of activities and shows that approximately 50 % of all cases follow a single variant, emphasising the value of Pareto analysis【608708606862785†L114-L124】.
5. **Quality Dimensions** – the healthcare study details how PM4Py computes the five quality metrics (fitness, precision, generalisation, simplicity and soundness) using token‑based replay and alignments【378107908900392†L1170-L1277】.
6. **Performance Analysis** – the same study demonstrates performance analysis by measuring event distributions over time, throughput times, case arrival ratios and sojourn times, and shows how these metrics reveal bottlenecks and resource constraints【378107908900392†L1318-L1424】【378107908900392†L1472-L1483】.
7. **PM4Py Functionality** – the PM4Py architecture paper lists the library’s features, including process discovery algorithms, conformance checking, filtering, case management, graphs and social network analysis (handover‑of‑work, working together, subcontracting)【943788946304099†L189-L204】.
8. **PM4Py Tutorial** – a tutorial on PM4Py demonstrates loading XES logs, discovering Petri nets with the Inductive Miner, calculating case durations, identifying start/end activities, filtering logs by time or performance, creating dotted charts and computing statistics such as events per case【969964019365071†L84-L91】【969964019365071†L109-L170】.
9. **Heuristic Miner Filtering** – an explanation of process mining algorithms notes that the Heuristic Miner applies filtering to remove infrequent behaviour, making the resulting models more robust in complex data environments【362502486590578†L240-L253】.
