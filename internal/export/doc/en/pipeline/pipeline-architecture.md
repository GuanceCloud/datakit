# Pipeline Basics and Principles
---

The following is the module design and working principle of the DataKit Pipeline, which can help you better understand the Pipeline function, but you can choose to skip the following content and start using it directly.

## Data Flow in DataKit {#data-flow}

After various DataKit collector plug-ins or DataKit API collect or receive data, the data will be processed by the Pipeline function and then uploaded.

DataKit Pipeline includes a programmable data processor (Pipeline) and a programmable [data filter](../../datakit/datakit-filter.md) (Filter), the data processor is used for data processing, filtering, etc., and the data Filters focus on data filtering functionality.

The data flow in the simplified DataKit is shown in the figure below:

![data-flow](img/pipeline-data-flow.drawio.png)

## Data Processor Workflow {#data-processor}

The workflow of the Pipeline data processor is shown in the data flow chart:

![data-processor](img/pipeline-data-processor.drawio.png)
