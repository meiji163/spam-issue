from matplotlib import pyplot as plt
import pandas as pd
from sklearn import tree

dataset = pd.read_csv("cli-cli_data.csv")

features = list(dataset.columns)[:-1] 
inps = dataset[features]
targs = dataset["is_spam"]

clf = tree.DecisionTreeClassifier(
        random_state=0, 
        criterion="entropy",
        class_weight="balanced"
    )

clf.fit(inps, targs)
clf.score(inps,targs)

fig = plt.figure(figsize=(25,20))
_ = tree.plot_tree(
        clf, 
        max_depth=2,
        filled=True,
        feature_names = features,
        class_names = ["spam","not spam"]
    )

